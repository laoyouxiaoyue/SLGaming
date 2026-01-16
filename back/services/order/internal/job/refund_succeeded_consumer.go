package job

import (
	"context"
	"encoding/json"

	pkgIoc "SLGaming/back/pkg/ioc"
	"SLGaming/back/services/order/internal/model"
	orderMQ "SLGaming/back/services/order/internal/mq"
	"SLGaming/back/services/order/internal/svc"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// refundSucceededEventPayload 与 user 服务发送的 ORDER_REFUND_SUCCEEDED 事件对应
type refundSucceededEventPayload struct {
	OrderID    uint64 `json:"order_id"`
	OrderNo    string `json:"order_no"`
	BizOrderID string `json:"biz_order_id"`
}

const eventTypeRefundSucceeded = "ORDER_REFUND_SUCCEEDED"

// StartRefundSucceededConsumer 启动消费 ORDER_REFUND_SUCCEEDED 事件的 Consumer
func StartRefundSucceededConsumer(ctx context.Context, svcCtx *svc.ServiceContext) {
	cfg := svcCtx.Config.RocketMQ
	if len(cfg.NameServers) == 0 {
		logx.Infof("refund succeeded consumer not started: rocketmq not configured")
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
		"order-refund-consumer",
		[]string{orderMQ.OrderEventTopic()},
		func(c context.Context, msg *primitive.MessageExt) error {
			// 只处理 ORDER_REFUND_SUCCEEDED 事件
			if msg.GetTags() != eventTypeRefundSucceeded {
				return nil
			}
			return handleRefundSucceeded(c, svcCtx.DB, msg)
		},
	)
	if err != nil {
		logx.Errorf("init refund succeeded consumer failed: %v", err)
		return
	}

	go func() {
		<-ctx.Done()
		pkgIoc.ShutdownRocketMQConsumer(consumer)
	}()

	logx.Infof("refund succeeded consumer started, topic=%s", orderMQ.OrderEventTopic())
}

func handleRefundSucceeded(ctx context.Context, db *gorm.DB, msg *primitive.MessageExt) error {
	var payload refundSucceededEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logx.Errorf("unmarshal ORDER_REFUND_SUCCEEDED payload failed: %v, body=%s", err, string(msg.Body))
		return nil
	}

	if payload.OrderID == 0 && payload.OrderNo == "" {
		logx.Errorf("invalid ORDER_REFUND_SUCCEEDED payload: order_id=0 and order_no empty")
		return nil
	}

	// 根据 order_id 或 order_no 更新订单状态：CANCEL_REFUNDING -> CANCELLED
	return db.Transaction(func(tx *gorm.DB) error {
		var o model.Order
		var err error
		if payload.OrderID != 0 {
			err = tx.Where("id = ?", payload.OrderID).First(&o).Error
		} else {
			err = tx.Where("order_no = ?", payload.OrderNo).First(&o).Error
		}

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				logx.Errorf("order not found when handling ORDER_REFUND_SUCCEEDED, order_id=%d, order_no=%s",
					payload.OrderID, payload.OrderNo)
				return nil
			}
			logx.Errorf("get order failed when handling ORDER_REFUND_SUCCEEDED: %v", err)
			return err
		}

		// 只有在取消中状态时才更新为已取消
		if o.Status != model.OrderStatusCancelRefunding {
			// 状态不符，直接返回，不视为错误，避免重复处理
			return nil
		}

		o.Status = model.OrderStatusCancelled
		if err := tx.Save(&o).Error; err != nil {
			logx.Errorf("update order status to CANCELLED failed: %v", err)
			return err
		}

		return nil
	})
}
