package job

import (
	"context"
	"encoding/json"
	"errors"

	pkgIoc "SLGaming/back/pkg/ioc"
	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/logic"
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	orderEventTopic           = "order_events"
	eventTypeOrderCancelled   = "ORDER_CANCELLED"
	eventTypeOrderCompleted   = "ORDER_COMPLETED"
	eventTypePaymentPending   = "ORDER_PAYMENT_PENDING"
	eventTypePaymentSucceeded = "ORDER_PAYMENT_SUCCEEDED"
	eventTypePaymentFailed    = "ORDER_PAYMENT_FAILED"
	eventTypeRefundSucceeded  = "ORDER_REFUND_SUCCEEDED"
)

// orderCancelledEventPayload 与订单服务中构造的 payload 对应
// 注意：扩展字段（NeedRefund, CancelReason）由订单服务使用，user 服务只使用基础字段
type orderCancelledEventPayload struct {
	OrderID     uint64 `json:"order_id"`
	OrderNo     string `json:"order_no"`
	BossID      uint64 `json:"boss_id"`
	CompanionID uint64 `json:"companion_id"`
	Amount      int64  `json:"amount"`
	BizOrderID  string `json:"biz_order_id"`

	// 扩展字段：用于订单服务的本地事务（user 服务忽略）
	NeedRefund   bool   `json:"need_refund"`
	CancelReason string `json:"cancel_reason"`
}

// orderCompletedEventPayload 订单完成事件负载（与订单服务中构造的 payload 对应）
type orderCompletedEventPayload struct {
	OrderID     uint64 `json:"order_id"`
	OrderNo     string `json:"order_no"`
	BossID      uint64 `json:"boss_id"`
	CompanionID uint64 `json:"companion_id"`
	Amount      int64  `json:"amount"`
	BizOrderID  string `json:"biz_order_id"`
}

// orderPaymentPendingEventPayload 订单支付待处理事件负载（与订单服务中构造的 payload 对应）
// 注意：扩展字段（CompanionID, GameName 等）由订单服务使用，user 服务只使用基础字段
type orderPaymentPendingEventPayload struct {
	OrderID    uint64 `json:"order_id"`
	OrderNo    string `json:"order_no"`
	BossID     uint64 `json:"boss_id"`
	Amount     int64  `json:"amount"`
	BizOrderID string `json:"biz_order_id"`

	// 扩展字段：用于订单服务的本地事务（user 服务忽略）
	CompanionID   uint64 `json:"companion_id"`
	GameName      string `json:"game_name"`
	DurationHours int32  `json:"duration_hours"`
	PricePerHour  int64  `json:"price_per_hour"`
}

// StartOrderRefundConsumer 启动消费订单取消事件的 RocketMQ Consumer
func StartOrderRefundConsumer(ctx context.Context, svcCtx *svc.ServiceContext) {
	cfg := svcCtx.Config().RocketMQ
	if len(cfg.NameServers) == 0 {
		helper.LogInfo(logx.WithContext(ctx), helper.OpMQConsumer, "order refund consumer not started: rocketmq not configured", nil)
		return
	}

	mqCfg := &pkgIoc.RocketMQConfigAdapter{
		NameServers: cfg.NameServers,
		Namespace:   cfg.Namespace,
		AccessKey:   cfg.AccessKey,
		SecretKey:   cfg.SecretKey,
	}

	consumer, err := pkgIoc.InitRocketMQConsumerWithSelector(
		mqCfg,
		"user-refund-consumer",
		[]string{orderEventTopic},
		eventTypeOrderCancelled+"||"+eventTypeOrderCompleted+"||"+eventTypePaymentPending+"||"+eventTypePaymentSucceeded+"||"+eventTypePaymentFailed+"||"+eventTypeRefundSucceeded,
		func(c context.Context, msg *primitive.MessageExt) error {
			return handleOrderEvent(c, svcCtx, msg)
		},
	)
	if err != nil {
		helper.LogError(logx.WithContext(ctx), helper.OpMQConsumer, "init order refund consumer failed", err, nil)
		return
	}

	// 确保进程退出时关闭 consumer
	go func() {
		<-ctx.Done()
		pkgIoc.ShutdownRocketMQConsumer(consumer)
	}()

	helper.LogSuccess(logx.WithContext(ctx), helper.OpMQConsumer, map[string]interface{}{
		"consumer": "order_refund",
		"topic":    orderEventTopic,
	})
}

func handleOrderEvent(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	eventType := msg.GetTags()
	switch eventType {
	case eventTypeOrderCancelled:
		return handleOrderCancelled(ctx, svcCtx, msg)
	case eventTypeOrderCompleted:
		return handleOrderCompleted(ctx, svcCtx, msg)
	case eventTypePaymentPending:
		return handlePaymentPending(ctx, svcCtx, msg)
	default:
		// 其他事件类型先忽略，后续可以扩展
		return nil
	}
}

// handleOrderCancelled 处理 ORDER_CANCELLED 事件，执行钱包退款
// 注意：只有当 NeedRefund=true 时才执行退款（已创建但未支付的订单不需要退款）
func handleOrderCancelled(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	logger := logx.WithContext(ctx)

	var payload orderCancelledEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "unmarshal ORDER_CANCELLED payload failed", err, map[string]interface{}{
			"body": string(msg.Body),
		})
		return nil // 丢弃这条，避免一直重试
	}

	if payload.BossID == 0 || payload.Amount <= 0 {
		helper.LogError(logger, helper.OpMQConsumer, "invalid ORDER_CANCELLED payload", nil, map[string]interface{}{
			"boss_id": payload.BossID,
			"amount":  payload.Amount,
		})
		return nil
	}

	// 只有当 NeedRefund=true 时才执行退款
	// 如果订单未支付（NeedRefund=false），则不需要退款，直接返回成功
	if !payload.NeedRefund {
		helper.LogInfo(logger, helper.OpMQConsumer, "order cancelled without refund (order was not paid)", map[string]interface{}{
			"order_no": payload.OrderNo,
			"boss_id":  payload.BossID,
		})
		return nil
	}

	// 使用退款逻辑进行幂等退款（Amount 为正数，BizOrderID 用于幂等控制），通过 RocketMQ 事务消息发送退款成功事件
	l := logic.NewRefundLogic(ctx, svcCtx)
	if err := l.Refund(payload.BossID, payload.Amount, payload.BizOrderID, payload.OrderNo, "order refund"); err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "refund wallet failed", err, map[string]interface{}{
			"order_no": payload.OrderNo,
			"boss_id":  payload.BossID,
			"amount":   payload.Amount,
		})
		return err
	}

	helper.LogSuccess(logger, helper.OpMQConsumer, map[string]interface{}{
		"event":    "order_cancelled",
		"order_no": payload.OrderNo,
		"boss_id":  payload.BossID,
		"amount":   payload.Amount,
		"refunded": true,
	})
	return nil
}

// handleOrderCompleted 处理 ORDER_COMPLETED 事件，给陪玩充值
func handleOrderCompleted(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	logger := logx.WithContext(ctx)

	var payload orderCompletedEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "unmarshal ORDER_COMPLETED payload failed", err, map[string]interface{}{
			"body": string(msg.Body),
		})
		return nil // 丢弃这条，避免一直重试
	}

	if payload.CompanionID == 0 || payload.Amount <= 0 {
		helper.LogError(logger, helper.OpMQConsumer, "invalid ORDER_COMPLETED payload", nil, map[string]interface{}{
			"companion_id": payload.CompanionID,
			"amount":       payload.Amount,
		})
		return nil
	}

	// 使用充值逻辑给陪玩加钱（Amount 为正数，BizOrderID 用于幂等控制）
	l := logic.NewRechargeLogic(ctx, svcCtx)
	_, err := l.Recharge(&user.RechargeRequest{
		UserId:     payload.CompanionID,
		Amount:     payload.Amount,
		BizOrderId: payload.BizOrderID,
		Remark:     "order completed payment",
	})
	if err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "recharge wallet failed", err, map[string]interface{}{
			"order_no":     payload.OrderNo,
			"companion_id": payload.CompanionID,
			"amount":       payload.Amount,
		})
		return err
	}

	helper.LogSuccess(logger, helper.OpMQConsumer, map[string]interface{}{
		"event":        "order_completed",
		"order_no":     payload.OrderNo,
		"companion_id": payload.CompanionID,
		"amount":       payload.Amount,
	})
	return nil
}

// handlePaymentPending 处理 ORDER_PAYMENT_PENDING 事件，执行扣款
// 扣款成功后，通过 RocketMQ 事务消息发送支付结果事件
func handlePaymentPending(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	logger := logx.WithContext(ctx)

	var payload orderPaymentPendingEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "unmarshal ORDER_PAYMENT_PENDING payload failed", err, map[string]interface{}{
			"body": string(msg.Body),
		})
		return nil // 丢弃这条，避免一直重试
	}

	if payload.BossID == 0 || payload.Amount <= 0 || payload.OrderNo == "" {
		helper.LogError(logger, helper.OpMQConsumer, "invalid ORDER_PAYMENT_PENDING payload", nil, map[string]interface{}{
			"boss_id":  payload.BossID,
			"amount":   payload.Amount,
			"order_no": payload.OrderNo,
		})
		return nil
	}

	db := svcCtx.DB().WithContext(ctx)
	var wallet model.UserWallet
	var paymentSucceeded bool
	var paymentFailedReason string

	// 扣款操作
	err := db.Transaction(func(tx *gorm.DB) error {
		// 0. 幂等检查：如果已经存在 CONSUME + biz_order_id 的流水，说明已经扣过款了
		var existedTr model.WalletTransaction
		if err := tx.
			Where("type = ? AND biz_order_id = ?", "CONSUME", payload.BizOrderID).
			First(&existedTr).Error; err == nil {
			// 已经扣过款了，直接返回成功（幂等）
			helper.LogInfo(logger, helper.OpMQConsumer, "payment already consumed, skip duplicate", map[string]interface{}{
				"order_no":     payload.OrderNo,
				"biz_order_id": payload.BizOrderID,
			})
			paymentSucceeded = true
			return nil
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// 1. 执行扣款逻辑（复制 Consume 的核心逻辑，但使用传入的 tx）
		// 加锁读取钱包记录，避免并发更新问题
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", payload.BossID).
			First(&wallet).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 钱包不存在，扣款失败
			paymentFailedReason = "wallet not found"
			return nil // 不返回错误，让事务提交
		} else if err != nil {
			return err
		}

		if wallet.Balance < payload.Amount {
			// 余额不足，扣款失败
			paymentFailedReason = "insufficient handsome coins"
			return nil // 不返回错误，让事务提交
		}

		// 执行扣款
		before := wallet.Balance
		after := before - payload.Amount
		wallet.Balance = after

		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		// 记录扣款流水
		tr := &model.WalletTransaction{
			UserID:        payload.BossID,
			WalletID:      wallet.ID,
			ChangeAmount:  -payload.Amount, // 消费为负数
			BeforeBalance: before,
			AfterBalance:  after,
			Type:          "CONSUME",
			BizOrderID:    payload.BizOrderID,
			Remark:        "order payment",
		}

		if err := tx.Create(tr).Error; err != nil {
			return err
		}

		// 扣款成功
		paymentSucceeded = true
		return nil
	})

	if err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "payment transaction failed", err, map[string]interface{}{
			"order_no": payload.OrderNo,
		})
		return err
	}

	// 事务提交后，发送支付结果事件（使用普通 Producer，非事务消息）
	if paymentSucceeded {
		return sendPaymentSucceededEvent(ctx, svcCtx, &payload)
	} else {
		return sendPaymentFailedEvent(ctx, svcCtx, &payload, paymentFailedReason)
	}
}

// sendPaymentSucceededEvent 发送支付成功事件（事务提交后调用）
func sendPaymentSucceededEvent(ctx context.Context, svcCtx *svc.ServiceContext, payload *orderPaymentPendingEventPayload) error {
	logger := logx.WithContext(ctx)

	if svcCtx.EventProducer == nil {
		helper.LogError(logger, helper.OpMQConsumer, "event producer not initialized, cannot send payment succeeded event", nil, nil)
		return nil // 不返回错误，避免影响主流程
	}

	succeededPayload := map[string]any{
		"order_id":     payload.OrderID,
		"order_no":     payload.OrderNo,
		"boss_id":      payload.BossID,
		"amount":       payload.Amount,
		"biz_order_id": payload.BizOrderID,
	}
	succeededPayloadJSON, err := json.Marshal(succeededPayload)
	if err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "marshal payment succeeded payload failed", err, nil)
		return nil // 不返回错误，避免影响主流程
	}

	msg := primitive.NewMessage(orderEventTopic, succeededPayloadJSON)
	msg.WithTag(eventTypePaymentSucceeded)

	if _, err := svcCtx.EventProducer.SendSync(ctx, msg); err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "send payment succeeded event failed", err, map[string]interface{}{
			"order_no": payload.OrderNo,
		})
		return nil // 不返回错误，避免影响主流程（消息发送失败不影响扣款结果）
	}

	helper.LogSuccess(logger, helper.OpMQConsumer, map[string]interface{}{
		"event":    "payment_succeeded",
		"order_no": payload.OrderNo,
		"boss_id":  payload.BossID,
		"amount":   payload.Amount,
	})
	return nil
}

// sendPaymentFailedEvent 发送支付失败事件（事务提交后调用）
func sendPaymentFailedEvent(ctx context.Context, svcCtx *svc.ServiceContext, payload *orderPaymentPendingEventPayload, reason string) error {
	logger := logx.WithContext(ctx)

	if svcCtx.EventProducer == nil {
		helper.LogError(logger, helper.OpMQConsumer, "event producer not initialized, cannot send payment failed event", nil, nil)
		return nil // 不返回错误，避免影响主流程
	}

	failedPayload := map[string]any{
		"order_id":     payload.OrderID,
		"order_no":     payload.OrderNo,
		"boss_id":      payload.BossID,
		"amount":       payload.Amount,
		"biz_order_id": payload.BizOrderID,
		"reason":       reason,
	}
	failedPayloadJSON, err := json.Marshal(failedPayload)
	if err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "marshal payment failed payload failed", err, nil)
		return nil // 不返回错误，避免影响主流程
	}

	msg := primitive.NewMessage(orderEventTopic, failedPayloadJSON)
	msg.WithTag(eventTypePaymentFailed)

	if _, err := svcCtx.EventProducer.SendSync(ctx, msg); err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "send payment failed event failed", err, map[string]interface{}{
			"order_no": payload.OrderNo,
		})
		return nil // 不返回错误，避免影响主流程
	}

	helper.LogError(logger, helper.OpMQConsumer, "consume wallet failed", nil, map[string]interface{}{
		"order_no": payload.OrderNo,
		"boss_id":  payload.BossID,
		"amount":   payload.Amount,
		"reason":   reason,
	})
	return nil
}
