package job

import (
	"context"
	"encoding/json"
	"time"

	pkgIoc "SLGaming/back/pkg/ioc"
	"SLGaming/back/services/order/internal/model"
	orderMQ "SLGaming/back/services/order/internal/mq"
	"SLGaming/back/services/order/internal/svc"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const (
	eventTypePaymentSucceeded = "ORDER_PAYMENT_SUCCEEDED"
	eventTypePaymentFailed    = "ORDER_PAYMENT_FAILED"
	eventTypeRefundSucceeded  = "ORDER_REFUND_SUCCEEDED"
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

// orderRefundSucceededEventPayload 订单退款成功事件负载（与用户服务中构造的 payload 对应）
type orderRefundSucceededEventPayload struct {
	OrderID    uint64 `json:"order_id"`
	OrderNo    string `json:"order_no"`
	BizOrderID string `json:"biz_order_id"`
	UserID     uint64 `json:"user_id"`
	Amount     int64  `json:"amount"`
}

// StartPaymentStatusConsumer 启动消费订单支付状态事件的 Consumer
func StartPaymentStatusConsumer(ctx context.Context, svcCtx *svc.ServiceContext) {
	cfg := svcCtx.Config.RocketMQ
	if len(cfg.NameServers) == 0 {
		logx.Infof("[payment_event] info: consumer not started, reason=rocketmq not configured")
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
		[]string{orderMQ.OrderEventTopic()},
		func(c context.Context, msg *primitive.MessageExt) error {
			return handlePaymentStatusEvent(c, svcCtx, msg)
		},
	)
	if err != nil {
		logx.Errorf("[payment_event] failed: init consumer failed, error=%v", err)
		return
	}

	go func() {
		<-ctx.Done()
		pkgIoc.ShutdownRocketMQConsumer(consumer)
	}()

	logx.Infof("[payment_event] succeeded: consumer started, topic=%s", orderMQ.OrderEventTopic())
}

func handlePaymentStatusEvent(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	eventType := msg.GetTags()
	switch eventType {
	case eventTypePaymentSucceeded:
		return handlePaymentSucceeded(ctx, svcCtx, msg)
	case eventTypePaymentFailed:
		return handlePaymentFailed(ctx, svcCtx, msg)
	case eventTypeRefundSucceeded:
		return handleRefundSucceeded(ctx, svcCtx, msg)
	default:
		return nil
	}
}

// handlePaymentSucceeded 处理 ORDER_PAYMENT_SUCCEEDED 事件，更新订单状态为 PAID
func handlePaymentSucceeded(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	var payload orderPaymentSucceededEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logx.Errorf("[payment_event] failed: unmarshal ORDER_PAYMENT_SUCCEEDED payload failed, body=%s, error=%v", string(msg.Body), err)
		return nil
	}

	if payload.OrderID == 0 && payload.OrderNo == "" {
		logx.Errorf("[payment_event] failed: invalid ORDER_PAYMENT_SUCCEEDED payload, order_id=0 and order_no empty")
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
				logx.Errorf("[payment_event] failed: order not found when handling ORDER_PAYMENT_SUCCEEDED, order_id=%d, order_no=%s", payload.OrderID, payload.OrderNo)
				return nil
			}
			logx.Errorf("[payment_event] failed: get order failed when handling ORDER_PAYMENT_SUCCEEDED, error=%v", err)
			return err
		}

		// 只有在 CREATED 状态时才更新为已支付（幂等）
		if o.Status != model.OrderStatusCreated {
			// 状态不符，直接返回，不视为错误，避免重复处理
			logx.Infof("[payment_event] warning: order status is not CREATED when handling ORDER_PAYMENT_SUCCEEDED, order_id=%d, status=%d", o.ID, o.Status)
			return nil
		}

		now := time.Now()
		o.Status = model.OrderStatusPaid
		o.PaidAt = &now

		if err := tx.Save(&o).Error; err != nil {
			logx.Errorf("[payment_event] failed: update order status to PAID failed, order_id=%d, error=%v", o.ID, err)
			return err
		}

		logx.Infof("[payment_event] succeeded: order payment succeeded, order_id=%d, order_no=%s", o.ID, o.OrderNo)
		return nil
	})
}

// handlePaymentFailed 处理 ORDER_PAYMENT_FAILED 事件，更新订单状态为 CANCELLED
func handlePaymentFailed(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	var payload orderPaymentFailedEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logx.Errorf("[payment_event] failed: unmarshal ORDER_PAYMENT_FAILED payload failed, body=%s, error=%v", string(msg.Body), err)
		return nil
	}

	if payload.OrderID == 0 && payload.OrderNo == "" {
		logx.Errorf("[payment_event] failed: invalid ORDER_PAYMENT_FAILED payload, order_id=0 and order_no empty")
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
				logx.Errorf("[payment_event] failed: order not found when handling ORDER_PAYMENT_FAILED, order_id=%d, order_no=%s", payload.OrderID, payload.OrderNo)
				return nil
			}
			logx.Errorf("[payment_event] failed: get order failed when handling ORDER_PAYMENT_FAILED, error=%v", err)
			return err
		}

		// 只有在 CREATED 状态时才更新为已取消（幂等）
		if o.Status != model.OrderStatusCreated {
			// 状态不符，直接返回，不视为错误，避免重复处理
			logx.Infof("[payment_event] warning: order status is not CREATED when handling ORDER_PAYMENT_FAILED, order_id=%d, status=%d", o.ID, o.Status)
			return nil
		}

		now := time.Now()
		o.Status = model.OrderStatusCancelled
		o.CancelledAt = &now
		o.CancelReason = "payment failed: " + payload.Reason

		if err := tx.Save(&o).Error; err != nil {
			logx.Errorf("[payment_event] failed: update order status to CANCELLED failed, order_id=%d, error=%v", o.ID, err)
			return err
		}

		logx.Infof("[payment_event] succeeded: order payment failed, order_id=%d, order_no=%s, reason=%s", o.ID, o.OrderNo, payload.Reason)
		return nil
	})
}

// handleRefundSucceeded 处理 ORDER_REFUND_SUCCEEDED 事件，更新订单状态为 CANCELLED
func handleRefundSucceeded(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	var payload orderRefundSucceededEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logx.Errorf("[payment_event] failed: unmarshal ORDER_REFUND_SUCCEEDED payload failed, body=%s, error=%v", string(msg.Body), err)
		return nil
	}

	if payload.OrderID == 0 && payload.OrderNo == "" {
		logx.Errorf("[payment_event] failed: invalid ORDER_REFUND_SUCCEEDED payload, order_id=0 and order_no empty")
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
				logx.Errorf("[payment_event] failed: order not found when handling ORDER_REFUND_SUCCEEDED, order_id=%d, order_no=%s", payload.OrderID, payload.OrderNo)
				return nil
			}
			logx.Errorf("[payment_event] failed: get order failed when handling ORDER_REFUND_SUCCEEDED, error=%v", err)
			return err
		}

		// 只有在 CANCEL_REFUNDING 状态时才更新为已取消（幂等）
		if o.Status != model.OrderStatusCancelRefunding {
			// 状态不符，直接返回，不视为错误，避免重复处理
			logx.Infof("[payment_event] warning: order status is not CANCEL_REFUNDING when handling ORDER_REFUND_SUCCEEDED, order_id=%d, status=%d", o.ID, o.Status)
			return nil
		}

		now := time.Now()
		o.Status = model.OrderStatusCancelled
		o.CancelledAt = &now

		if err := tx.Save(&o).Error; err != nil {
			logx.Errorf("[payment_event] failed: update order status to CANCELLED failed, order_id=%d, error=%v", o.ID, err)
			return err
		}

		logx.Infof("[payment_event] succeeded: order refund succeeded, order_id=%d, order_no=%s, amount=%d", o.ID, o.OrderNo, payload.Amount)
		return nil
	})
}
