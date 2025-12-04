package job

import (
	"context"
	"time"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const (
	userEventTopic          = "order_events" // 与订单服务使用同一 topic，区分靠 Tag
	userOutboxStatusPending = "PENDING"
	userOutboxStatusSent    = "SENT"
	userOutboxStatusFailed  = "FAILED"
	userOutboxBatchSize     = 50
	userOutboxInterval      = 2 * time.Second
)

// StartUserOutboxDispatcher 启动一个后台协程，定期扫描用户 Outbox 表并将事件发送到 RocketMQ
func StartUserOutboxDispatcher(ctx context.Context, svcCtx *svc.ServiceContext) {
	if svcCtx == nil || svcCtx.DB() == nil || svcCtx.EventProducer == nil {
		logx.Infof("user outbox dispatcher not started: db or producer is nil")
		return
	}

	go func() {
		logx.Infof("user outbox dispatcher started")
		ticker := time.NewTicker(userOutboxInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logx.Infof("user outbox dispatcher stopped: context canceled")
				return
			case <-ticker.C:
				if err := dispatchUserOutboxEvents(ctx, svcCtx.DB(), svcCtx.EventProducer); err != nil {
					logx.Errorf("dispatch user outbox events failed: %v", err)
				}
			}
		}
	}()
}

func dispatchUserOutboxEvents(ctx context.Context, db *gorm.DB, producer rocketmq.Producer) error {
	var events []model.UserEventOutbox
	if err := db.
		Where("status = ?", userOutboxStatusPending).
		Order("created_at ASC").
		Limit(userOutboxBatchSize).
		Find(&events).Error; err != nil {
		return err
	}

	if len(events) == 0 {
		return nil
	}

	for _, evt := range events {
		if evt.Payload == "" {
			_ = markUserEventFailed(db, &evt, "empty payload")
			continue
		}
		msg := primitive.NewMessage(userEventTopic, []byte(evt.Payload))
		msg.WithTag(evt.EventType)

		if _, err := producer.SendSync(ctx, msg); err != nil {
			logx.Errorf("send user event message failed: id=%d, type=%s, err=%v", evt.ID, evt.EventType, err)
			_ = markUserEventFailed(db, &evt, err.Error())
			continue
		}

		if err := markUserEventSent(db, &evt); err != nil {
			logx.Errorf("mark user event outbox sent failed: id=%d, err=%v", evt.ID, err)
		}
	}

	return nil
}

func markUserEventSent(db *gorm.DB, evt *model.UserEventOutbox) error {
	return db.Model(evt).Updates(map[string]any{
		"status":     userOutboxStatusSent,
		"last_error": "",
	}).Error
}

func markUserEventFailed(db *gorm.DB, evt *model.UserEventOutbox, errMsg string) error {
	return db.Model(evt).Updates(map[string]any{
		"status":     userOutboxStatusFailed,
		"last_error": errMsg,
	}).Error
}
