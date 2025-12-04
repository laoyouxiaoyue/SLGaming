package job

import (
	"context"
	"encoding/json"
	"errors"

	pkgIoc "SLGaming/back/pkg/ioc"
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
type orderCancelledEventPayload struct {
	OrderID     uint64 `json:"order_id"`
	OrderNo     string `json:"order_no"`
	BossID      uint64 `json:"boss_id"`
	CompanionID uint64 `json:"companion_id"`
	Amount      int64  `json:"amount"`
	BizOrderID  string `json:"biz_order_id"`
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
type orderPaymentPendingEventPayload struct {
	OrderID    uint64 `json:"order_id"`
	OrderNo    string `json:"order_no"`
	BossID     uint64 `json:"boss_id"`
	Amount     int64  `json:"amount"`
	BizOrderID string `json:"biz_order_id"`
}

// StartOrderRefundConsumer 启动消费订单取消事件的 RocketMQ Consumer
func StartOrderRefundConsumer(ctx context.Context, svcCtx *svc.ServiceContext) {
	cfg := svcCtx.Config().RocketMQ
	if len(cfg.NameServers) == 0 {
		logx.Infof("order refund consumer not started: rocketmq not configured")
		return
	}

	mqCfg := &pkgIoc.RocketMQConfigAdapter{
		NameServers: cfg.NameServers,
		Namespace:   cfg.Namespace,
		AccessKey:   cfg.AccessKey,
		SecretKey:   cfg.SecretKey,
	}

	consumer, err := pkgIoc.InitRocketMQConsumer(
		mqCfg,
		"user-refund-consumer",
		[]string{orderEventTopic},
		func(c context.Context, msg *primitive.MessageExt) error {
			return handleOrderEvent(c, svcCtx, msg)
		},
	)
	if err != nil {
		logx.Errorf("init order refund consumer failed: %v", err)
		return
	}

	// 确保进程退出时关闭 consumer
	go func() {
		<-ctx.Done()
		pkgIoc.ShutdownRocketMQConsumer(consumer)
	}()

	logx.Infof("order refund consumer started, topic=%s", orderEventTopic)
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
func handleOrderCancelled(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	var payload orderCancelledEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logx.Errorf("unmarshal ORDER_CANCELLED payload failed: %v, body=%s", err, string(msg.Body))
		return nil // 丢弃这条，避免一直重试
	}

	if payload.BossID == 0 || payload.Amount <= 0 {
		logx.Errorf("invalid ORDER_CANCELLED payload: boss_id=%d, amount=%d", payload.BossID, payload.Amount)
		return nil
	}

	// 使用退款逻辑进行幂等退款（Amount 为正数，BizOrderID 用于幂等控制），并写入 Outbox
	l := logic.NewRefundLogic(ctx, svcCtx)
	if err := l.Refund(payload.BossID, payload.Amount, payload.BizOrderID, payload.OrderNo, "order refund"); err != nil {
		logx.Errorf("refund wallet failed for order %s, user=%d, amount=%d, err=%v",
			payload.OrderNo, payload.BossID, payload.Amount, err)
		return err
	}

	return nil
}

// handleOrderCompleted 处理 ORDER_COMPLETED 事件，给陪玩充值
func handleOrderCompleted(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	var payload orderCompletedEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logx.Errorf("unmarshal ORDER_COMPLETED payload failed: %v, body=%s", err, string(msg.Body))
		return nil // 丢弃这条，避免一直重试
	}

	if payload.CompanionID == 0 || payload.Amount <= 0 {
		logx.Errorf("invalid ORDER_COMPLETED payload: companion_id=%d, amount=%d", payload.CompanionID, payload.Amount)
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
		logx.Errorf("recharge wallet failed for order %s, companion=%d, amount=%d, err=%v",
			payload.OrderNo, payload.CompanionID, payload.Amount, err)
		return err
	}

	logx.Infof("recharge wallet success for order %s, companion=%d, amount=%d",
		payload.OrderNo, payload.CompanionID, payload.Amount)
	return nil
}

// handlePaymentPending 处理 ORDER_PAYMENT_PENDING 事件，执行扣款
// 在一个事务中：扣款 + 写入支付结果事件到 Outbox，保证原子性
func handlePaymentPending(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	var payload orderPaymentPendingEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logx.Errorf("unmarshal ORDER_PAYMENT_PENDING payload failed: %v, body=%s", err, string(msg.Body))
		return nil // 丢弃这条，避免一直重试
	}

	if payload.BossID == 0 || payload.Amount <= 0 || payload.OrderNo == "" {
		logx.Errorf("invalid ORDER_PAYMENT_PENDING payload: boss_id=%d, amount=%d, order_no=%s",
			payload.BossID, payload.Amount, payload.OrderNo)
		return nil
	}

	db := svcCtx.DB().WithContext(ctx)
	var wallet model.UserWallet

	// 在一个事务中：扣款 + 写入支付结果事件到 Outbox，保证原子性
	return db.Transaction(func(tx *gorm.DB) error {
		// 0. 幂等检查：如果已经存在 CONSUME + biz_order_id 的流水，说明已经扣过款了
		var existedTr model.WalletTransaction
		if err := tx.
			Where("type = ? AND biz_order_id = ?", "CONSUME", payload.BizOrderID).
			First(&existedTr).Error; err == nil {
			// 已经扣过款了，直接写入成功事件（幂等）
			logx.Infof("payment already consumed for order %s, biz_order_id=%s, skip duplicate consume",
				payload.OrderNo, payload.BizOrderID)
			return writePaymentSucceededEvent(tx, &payload)
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// 1. 执行扣款逻辑（复制 Consume 的核心逻辑，但使用传入的 tx）
		// 加锁读取钱包记录，避免并发更新问题
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", payload.BossID).
			First(&wallet).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 钱包不存在，扣款失败，写入失败事件（不返回错误，让事务提交）
			return writePaymentFailedEvent(tx, &payload, "wallet not found")
		} else if err != nil {
			return err
		}

		if wallet.Balance < payload.Amount {
			// 余额不足，扣款失败，写入失败事件（不返回错误，让事务提交）
			return writePaymentFailedEvent(tx, &payload, "insufficient handsome coins")
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

		// 2. 扣款成功，写入 ORDER_PAYMENT_SUCCEEDED 事件到 Outbox（同一事务）
		return writePaymentSucceededEvent(tx, &payload)
	})
}

// writePaymentSucceededEvent 写入支付成功事件到 Outbox（在事务中调用）
func writePaymentSucceededEvent(tx *gorm.DB, payload *orderPaymentPendingEventPayload) error {
	succeededPayload := map[string]any{
		"order_id":     payload.OrderID,
		"order_no":     payload.OrderNo,
		"boss_id":      payload.BossID,
		"amount":       payload.Amount,
		"biz_order_id": payload.BizOrderID,
	}
	succeededPayloadJSON, err := json.Marshal(succeededPayload)
	if err != nil {
		logx.Errorf("marshal payment succeeded payload failed: %v", err)
		return err
	}

	evt := &model.UserEventOutbox{
		EventType: eventTypePaymentSucceeded,
		Payload:   string(succeededPayloadJSON),
		Status:    "PENDING",
	}
	if err := tx.Create(evt).Error; err != nil {
		logx.Errorf("create payment succeeded event outbox failed: %v", err)
		return err
	}

	logx.Infof("payment succeeded for order %s, boss=%d, amount=%d",
		payload.OrderNo, payload.BossID, payload.Amount)
	return nil
}

// writePaymentFailedEvent 写入支付失败事件到 Outbox（在事务中调用）
func writePaymentFailedEvent(tx *gorm.DB, payload *orderPaymentPendingEventPayload, reason string) error {
	failedPayload := map[string]any{
		"order_id":     payload.OrderID,
		"order_no":     payload.OrderNo,
		"boss_id":      payload.BossID,
		"amount":       payload.Amount,
		"biz_order_id": payload.BizOrderID,
		"reason":       reason,
	}
	failedPayloadJSON, _ := json.Marshal(failedPayload)

	evt := &model.UserEventOutbox{
		EventType: eventTypePaymentFailed,
		Payload:   string(failedPayloadJSON),
		Status:    "PENDING",
	}
	if err := tx.Create(evt).Error; err != nil {
		logx.Errorf("create payment failed event outbox failed: %v", err)
		return err
	}

	logx.Errorf("consume wallet failed for order %s, boss=%d, amount=%d, reason=%s",
		payload.OrderNo, payload.BossID, payload.Amount, reason)
	return nil
}
