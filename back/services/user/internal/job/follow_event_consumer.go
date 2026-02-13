package job

import (
	"context"
	"encoding/json"

	pkgIoc "SLGaming/back/pkg/ioc"
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/mq"
	"SLGaming/back/services/user/internal/svc"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// StartFollowEventConsumer 启动关注和取消关注事件的消费者
func StartFollowEventConsumer(ctx context.Context, svcCtx *svc.ServiceContext) {
	cfg := svcCtx.Config().RocketMQ
	if len(cfg.NameServers) == 0 {
		logx.Infof("follow event consumer not started: rocketmq not configured")
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
		"user-follow-consumer",
		[]string{mq.FollowEventTopic()},
		func(c context.Context, msg *primitive.MessageExt) error {
			return handleFollowEvent(c, svcCtx, msg)
		},
	)
	if err != nil {
		logx.Errorf("init follow event consumer failed: %v", err)
		return
	}

	// 确保进程退出时关闭 consumer
	go func() {
		<-ctx.Done()
		pkgIoc.ShutdownRocketMQConsumer(consumer)
	}()

	logx.Infof("follow event consumer started, topic=%s", mq.FollowEventTopic())
}

// handleFollowEvent 处理关注和取消关注事件
func handleFollowEvent(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	if msg == nil {
		return nil
	}

	msgID := msg.MsgId
	db := svcCtx.DB().WithContext(ctx)

	// 幂等性检查：检查消息是否已经被处理
	var processedMsg model.ProcessedMessage
	if err := db.Where("message_id = ?", msgID).First(&processedMsg).Error; err == nil {
		// 消息已经被处理，直接返回成功
		logx.Infof("follow event message already processed: %s", msgID)
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	var handleErr error
	switch msg.GetTags() {
	case mq.EventTypeFollowUser():
		handleErr = handleFollowUser(ctx, db, msg)
	case mq.EventTypeUnfollowUser():
		handleErr = handleUnfollowUser(ctx, db, msg)
	default:
		// 忽略其他事件类型
		return nil
	}

	if handleErr != nil {
		return handleErr
	}

	// 处理成功后记录消息ID
	processedMsg = model.ProcessedMessage{
		MessageID: msgID,
		EventType: msg.GetTags(),
	}
	if err := db.Create(&processedMsg).Error; err != nil {
		logx.Errorf("failed to record processed follow event message: %v", err)
		// 记录失败不影响主流程，因为消息处理已经成功
	}

	return nil
}

// handleFollowUser 处理关注用户事件
func handleFollowUser(ctx context.Context, db *gorm.DB, msg *primitive.MessageExt) error {
	var payload mq.FollowUserPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logx.Errorf("unmarshal follow user payload failed: %v, body=%s", err, string(msg.Body))
		return nil // 丢弃这条，避免一直重试
	}

	if payload.FollowerID == 0 || payload.FollowingID == 0 {
		logx.Errorf("invalid follow user payload: follower_id=%d, following_id=%d", payload.FollowerID, payload.FollowingID)
		return nil
	}

	// 在本地事务中更新用户计数
	return db.Transaction(func(tx *gorm.DB) error {
		// 增加被关注用户的粉丝数
		if err := tx.Model(&model.User{}).Where("id = ?", payload.FollowingID).
			UpdateColumn("follower_count", gorm.Expr("follower_count + ?", 1)).Error; err != nil {
			return err
		}

		// 增加关注者的关注数
		if err := tx.Model(&model.User{}).Where("id = ?", payload.FollowerID).
			UpdateColumn("following_count", gorm.Expr("following_count + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})
}

// handleUnfollowUser 处理取消关注用户事件
func handleUnfollowUser(ctx context.Context, db *gorm.DB, msg *primitive.MessageExt) error {
	var payload mq.UnfollowUserPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logx.Errorf("unmarshal unfollow user payload failed: %v, body=%s", err, string(msg.Body))
		return nil // 丢弃这条，避免一直重试
	}

	if payload.FollowerID == 0 || payload.FollowingID == 0 {
		logx.Errorf("invalid unfollow user payload: follower_id=%d, following_id=%d", payload.FollowerID, payload.FollowingID)
		return nil
	}

	// 在本地事务中更新用户计数
	return db.Transaction(func(tx *gorm.DB) error {
		// 减少被关注用户的粉丝数
		if err := tx.Model(&model.User{}).Where("id = ?", payload.FollowingID).
			UpdateColumn("follower_count", gorm.Expr("GREATEST(follower_count - ?, 0)", 1)).Error; err != nil {
			return err
		}

		// 减少关注者的关注数
		if err := tx.Model(&model.User{}).Where("id = ?", payload.FollowerID).
			UpdateColumn("following_count", gorm.Expr("GREATEST(following_count - ?, 0)", 1)).Error; err != nil {
			return err
		}

		return nil
	})
}
