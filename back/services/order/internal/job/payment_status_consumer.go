package job

import (
	"context"
	"encoding/json"
	"time"

	pkgIoc "SLGaming/back/pkg/ioc"
	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/internal/svc"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const (
	eventTypePaymentSucceeded = "ORDER_PAYMENT_SUCCEEDED"
	eventTypePaymentFailed    = "ORDER_PAYMENT_FAILED"
)

// orderPaymentSucceededEventPayload 订单支付成功事件负载（与用户服务中构造的 payload 对应）
type orderPaymentSucceededEventPayload struct {
	OrderID    uint64 `json:"order_id"`
	OrderNo    string `json:"order_no"`
	BossID     uint64 `json:"boss_id"`
	Amount     int64  `json:"amount"`
	BizOrderID string `json:"biz_order_id"`
}

// orderPaymentFailedEventPayload 订单支付失败事件负载（与用户服务中构造的 payload 对应）
type orderPaymentFailedEventPayload struct {
	OrderID    uint64 `json:"order_id"`
	OrderNo    string `json:"order_no"`
	BossID     uint64 `json:"boss_id"`
	Amount     int64  `json:"amount"`
	BizOrderID string `json:"biz_order_id"`
	Reason     string `json:"reason"`
}

// StartPaymentStatusConsumer 启动消费订单支付状态事件的 Consumer
func StartPaymentStatusConsumer(ctx context.Context, svcCtx *svc.ServiceContext) {
	cfg := svcCtx.Config.RocketMQ
	if len(cfg.NameServers) == 0 {
		logx.Infof("payment status consumer not started: rocketmq not configured")
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
		"order-payment-status-consumer",
		[]string{orderEventTopic},
		func(c context.Context, msg *primitive.MessageExt) error {
			return handlePaymentStatusEvent(c, svcCtx, msg)
		},
	)
	if err != nil {
		logx.Errorf("init payment status consumer failed: %v", err)
		return
	}

	go func() {
		<-ctx.Done()
		pkgIoc.ShutdownRocketMQConsumer(consumer)
	}()

	logx.Infof("payment status consumer started, topic=%s", orderEventTopic)
}

func handlePaymentStatusEvent(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	eventType := msg.GetTags()
	switch eventType {
	case eventTypePaymentSucceeded:
		return handlePaymentSucceeded(ctx, svcCtx, msg)
	case eventTypePaymentFailed:
		return handlePaymentFailed(ctx, svcCtx, msg)
	default:
		return nil
	}
}

// handlePaymentSucceeded 处理 ORDER_PAYMENT_SUCCEEDED 事件，更新订单状态为 PAID
func handlePaymentSucceeded(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	var payload orderPaymentSucceededEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logx.Errorf("unmarshal ORDER_PAYMENT_SUCCEEDED payload failed: %v, body=%s", err, string(msg.Body))
		return nil
	}

	if payload.OrderID == 0 && payload.OrderNo == "" {
		logx.Errorf("invalid ORDER_PAYMENT_SUCCEEDED payload: order_id=0 and order_no empty")
		return nil
	}

	// 在一个事务中更新订单状态为已支付
	return svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var o model.Order
		var err error
		if payload.OrderID != 0 {
			err = tx.Where("id = ?", payload.OrderID).First(&o).Error
		} else {
			err = tx.Where("order_no = ?", payload.OrderNo).First(&o).Error
		}

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				logx.Errorf("order not found when handling ORDER_PAYMENT_SUCCEEDED, order_id=%d, order_no=%s",
					payload.OrderID, payload.OrderNo)
				return nil
			}
			logx.Errorf("get order failed when handling ORDER_PAYMENT_SUCCEEDED: %v", err)
			return err
		}

		// 只有在 CREATED 状态时才更新为已支付（幂等）
		if o.Status != model.OrderStatusCreated {
			// 状态不符，直接返回，不视为错误，避免重复处理
			logx.Infof("order status is not CREATED when handling ORDER_PAYMENT_SUCCEEDED, order_id=%d, status=%d",
				o.ID, o.Status)
			return nil
		}

		now := time.Now()
		o.Status = model.OrderStatusPaid
		o.PaidAt = &now

		if err := tx.Save(&o).Error; err != nil {
			logx.Errorf("update order status to PAID failed: %v", err)
			return err
		}

		logx.Infof("order payment succeeded, order_id=%d, order_no=%s", o.ID, o.OrderNo)
		return nil
	})
}

// handlePaymentFailed 处理 ORDER_PAYMENT_FAILED 事件，更新订单状态为 CANCELLED
func handlePaymentFailed(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	var payload orderPaymentFailedEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logx.Errorf("unmarshal ORDER_PAYMENT_FAILED payload failed: %v, body=%s", err, string(msg.Body))
		return nil
	}

	if payload.OrderID == 0 && payload.OrderNo == "" {
		logx.Errorf("invalid ORDER_PAYMENT_FAILED payload: order_id=0 and order_no empty")
		return nil
	}

	// 在一个事务中更新订单状态为已取消
	return svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var o model.Order
		var err error
		if payload.OrderID != 0 {
			err = tx.Where("id = ?", payload.OrderID).First(&o).Error
		} else {
			err = tx.Where("order_no = ?", payload.OrderNo).First(&o).Error
		}

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				logx.Errorf("order not found when handling ORDER_PAYMENT_FAILED, order_id=%d, order_no=%s",
					payload.OrderID, payload.OrderNo)
				return nil
			}
			logx.Errorf("get order failed when handling ORDER_PAYMENT_FAILED: %v", err)
			return err
		}

		// 只有在 CREATED 状态时才更新为已取消（幂等）
		if o.Status != model.OrderStatusCreated {
			// 状态不符，直接返回，不视为错误，避免重复处理
			logx.Infof("order status is not CREATED when handling ORDER_PAYMENT_FAILED, order_id=%d, status=%d",
				o.ID, o.Status)
			return nil
		}

		now := time.Now()
		o.Status = model.OrderStatusCancelled
		o.CancelledAt = &now
		o.CancelReason = "payment failed: " + payload.Reason

		if err := tx.Save(&o).Error; err != nil {
			logx.Errorf("update order status to CANCELLED failed: %v", err)
			return err
		}

		logx.Infof("order payment failed, order_id=%d, order_no=%s, reason=%s", o.ID, o.OrderNo, payload.Reason)
		return nil
	})
}
