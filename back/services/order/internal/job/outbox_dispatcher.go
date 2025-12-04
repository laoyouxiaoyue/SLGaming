package job

import (
	"context"
	"time"

	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/internal/svc"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const (
	orderEventTopic        = "order_events"
	outboxStatusPending    = "PENDING"
	outboxStatusSent       = "SENT"
	outboxStatusFailed     = "FAILED"
	outboxDispatchBatch    = 50
	outboxDispatchInterval = 2 * time.Second
)

// StartOutboxDispatcher 启动一个后台协程，定期扫描 Outbox 表并将事件发送到 RocketMQ
func StartOutboxDispatcher(ctx context.Context, svcCtx *svc.ServiceContext) {
	if svcCtx == nil || svcCtx.DB == nil || svcCtx.OrderEventProducer == nil {
		logx.Infof("outbox dispatcher not started: db or producer is nil")
		return
	}

	go func() {
		logx.Infof("order outbox dispatcher started")
		ticker := time.NewTicker(outboxDispatchInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logx.Infof("order outbox dispatcher stopped: context canceled")
				return
			case <-ticker.C:
				if err := dispatchPendingEvents(ctx, svcCtx.DB, svcCtx.OrderEventProducer); err != nil {
					logx.Errorf("dispatch outbox events failed: %v", err)
				}
			}
		}
	}()
}

func dispatchPendingEvents(ctx context.Context, db *gorm.DB, producer rocketmq.Producer) error {
	var events []model.OrderEventOutbox
	if err := db.
		Where("status = ?", outboxStatusPending).
		Order("created_at ASC").
		Limit(outboxDispatchBatch).
		Find(&events).Error; err != nil {
		return err
	}

	if len(events) == 0 {
		return nil
	}

	for _, evt := range events {
		if evt.Payload == "" {
			// 标记为失败，避免死循环
			_ = markEventFailed(db, &evt, "empty payload")
			continue
		}

		msg := primitive.NewMessage(orderEventTopic, []byte(evt.Payload))
		// 将事件类型放到 Tag 中，便于消费者按类型区分
		msg.WithTag(evt.EventType)

		_, err := producer.SendSync(ctx, msg)
		if err != nil {
			logx.Errorf("send rocketmq message failed, event_id=%d, type=%s, err=%v", evt.ID, evt.EventType, err)
			_ = markEventFailed(db, &evt, err.Error())
			continue
		}

		if err := markEventSent(db, &evt); err != nil {
			logx.Errorf("mark outbox event sent failed, event_id=%d, err=%v", evt.ID, err)
		}
	}

	return nil
}

func markEventSent(db *gorm.DB, evt *model.OrderEventOutbox) error {
	return db.Model(evt).Updates(map[string]any{
		"status":     outboxStatusSent,
		"last_error": "",
	}).Error
}

func markEventFailed(db *gorm.DB, evt *model.OrderEventOutbox, errMsg string) error {
	return db.Model(evt).Updates(map[string]any{
		"status":     outboxStatusFailed,
		"last_error": errMsg,
	}).Error
}
