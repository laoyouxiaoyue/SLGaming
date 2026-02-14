package job

import (
	"context"
	"encoding/json"

	pkgIoc "SLGaming/back/pkg/ioc"
	"SLGaming/back/services/user/internal/helper"
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
		helper.LogInfo(logx.WithContext(ctx), helper.OpMQConsumer, "follow event consumer not started: rocketmq not configured", nil)
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
		helper.LogError(logx.WithContext(ctx), helper.OpMQConsumer, "init follow event consumer failed", err, nil)
		return
	}

	// 确保进程退出时关闭 consumer
	go func() {
		<-ctx.Done()
		pkgIoc.ShutdownRocketMQConsumer(consumer)
	}()

	helper.LogSuccess(logx.WithContext(ctx), helper.OpMQConsumer, map[string]interface{}{
		"consumer": "follow_event",
		"topic":    mq.FollowEventTopic(),
	})
}

// handleFollowEvent 处理关注和取消关注事件
func handleFollowEvent(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	logger := logx.WithContext(ctx)

	if msg == nil {
		return nil
	}

	msgID := msg.MsgId
	db := svcCtx.DB().WithContext(ctx)

	// 幂等性检查：检查消息是否已经被处理
	var processedMsg model.ProcessedMessage
	if err := db.Where("message_id = ?", msgID).First(&processedMsg).Error; err == nil {
		// 消息已经被处理，直接返回成功
		helper.LogInfo(logger, helper.OpMQConsumer, "message already processed, skip", map[string]interface{}{
			"message_id": msgID,
		})
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
		helper.LogError(logger, helper.OpMQConsumer, "record processed message failed", err, map[string]interface{}{
			"message_id": msgID,
		})
		// 记录失败不影响主流程，因为消息处理已经成功
	}

	return nil
}

// handleFollowUser 处理关注用户事件
func handleFollowUser(ctx context.Context, db *gorm.DB, msg *primitive.MessageExt) error {
	logger := logx.WithContext(ctx)

	var payload mq.FollowUserPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "unmarshal follow user payload failed", err, map[string]interface{}{
			"body": string(msg.Body),
		})
		return nil // 丢弃这条，避免一直重试
	}

	if payload.FollowerID == 0 || payload.FollowingID == 0 {
		helper.LogError(logger, helper.OpMQConsumer, "invalid follow user payload", nil, map[string]interface{}{
			"follower_id":  payload.FollowerID,
			"following_id": payload.FollowingID,
		})
		return nil
	}

	helper.LogInfo(logger, helper.OpMQConsumer, "processing follow user event", map[string]interface{}{
		"follower_id":  payload.FollowerID,
		"following_id": payload.FollowingID,
	})

	// 在本地事务中更新用户计数
	err := db.Transaction(func(tx *gorm.DB) error {
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

	if err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "update follow count failed", err, map[string]interface{}{
			"follower_id":  payload.FollowerID,
			"following_id": payload.FollowingID,
		})
		return err
	}

	helper.LogSuccess(logger, helper.OpMQConsumer, map[string]interface{}{
		"event":        "follow_user",
		"follower_id":  payload.FollowerID,
		"following_id": payload.FollowingID,
	})
	return nil
}

// handleUnfollowUser 处理取消关注用户事件
func handleUnfollowUser(ctx context.Context, db *gorm.DB, msg *primitive.MessageExt) error {
	logger := logx.WithContext(ctx)

	var payload mq.UnfollowUserPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "unmarshal unfollow user payload failed", err, map[string]interface{}{
			"body": string(msg.Body),
		})
		return nil // 丢弃这条，避免一直重试
	}

	if payload.FollowerID == 0 || payload.FollowingID == 0 {
		helper.LogError(logger, helper.OpMQConsumer, "invalid unfollow user payload", nil, map[string]interface{}{
			"follower_id":  payload.FollowerID,
			"following_id": payload.FollowingID,
		})
		return nil
	}

	helper.LogInfo(logger, helper.OpMQConsumer, "processing unfollow user event", map[string]interface{}{
		"follower_id":  payload.FollowerID,
		"following_id": payload.FollowingID,
	})

	// 在本地事务中更新用户计数
	err := db.Transaction(func(tx *gorm.DB) error {
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

	if err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "update unfollow count failed", err, map[string]interface{}{
			"follower_id":  payload.FollowerID,
			"following_id": payload.FollowingID,
		})
		return err
	}

	helper.LogSuccess(logger, helper.OpMQConsumer, map[string]interface{}{
		"event":        "unfollow_user",
		"follower_id":  payload.FollowerID,
		"following_id": payload.FollowingID,
	})
	return nil
}
