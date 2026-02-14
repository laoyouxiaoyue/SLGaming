package job

import (
	"context"
	"encoding/json"
	"time"

	"SLGaming/back/pkg/avatarmq"
	pkgIoc "SLGaming/back/pkg/ioc"
	"SLGaming/back/services/agent/agent"
	"SLGaming/back/services/agent/agentclient"
	"SLGaming/back/services/user/internal/config"
	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StartAvatarModerationConsumer 启动头像审核异步 Consumer
func StartAvatarModerationConsumer(ctx context.Context, svcCtx *svc.ServiceContext) {
	cfg := svcCtx.Config().RocketMQ
	if len(cfg.NameServers) == 0 {
		helper.LogInfo(logx.WithContext(ctx), helper.OpMQConsumer, "avatar moderation consumer not started: rocketmq not configured", nil)
		return
	}

	consumer, err := initAvatarConsumer(cfg, svcCtx)
	if err != nil {
		helper.LogError(logx.WithContext(ctx), helper.OpMQConsumer, "init avatar moderation consumer failed", err, nil)
		return
	}

	go func() {
		<-ctx.Done()
		shutdownAvatarConsumer(consumer)
	}()

	helper.LogSuccess(logx.WithContext(ctx), helper.OpMQConsumer, map[string]interface{}{
		"consumer": "avatar_moderation",
		"topic":    avatarmq.AvatarEventTopic(),
	})
}

func initAvatarConsumer(cfg config.RocketMQConf, svcCtx *svc.ServiceContext) (rocketmq.PushConsumer, error) {
	mqCfg := &pkgIoc.RocketMQConfigAdapter{
		NameServers: cfg.NameServers,
		Namespace:   cfg.Namespace,
		AccessKey:   cfg.AccessKey,
		SecretKey:   cfg.SecretKey,
	}
	return pkgIoc.InitRocketMQConsumer(
		mqCfg,
		"user-avatar-moderation-consumer",
		[]string{avatarmq.AvatarEventTopic()},
		func(c context.Context, msg *primitive.MessageExt) error {
			return handleAvatarEvent(c, svcCtx, msg)
		},
	)
}

func shutdownAvatarConsumer(consumer rocketmq.PushConsumer) {
	pkgIoc.ShutdownRocketMQConsumer(consumer)
}

func handleAvatarEvent(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	logger := logx.WithContext(ctx)

	if msg.GetTags() != avatarmq.EventTypeAvatarSubmit() {
		helper.LogInfo(logger, helper.OpMQConsumer, "skipping message with unexpected tag", map[string]interface{}{
			"tag":      msg.GetTags(),
			"expected": avatarmq.EventTypeAvatarSubmit(),
		})
		return nil
	}

	var payload avatarmq.AvatarModerationPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "unmarshal AvatarModeration payload failed", err, map[string]interface{}{
			"message_id": msg.MsgId,
			"body":       string(msg.Body),
		})
		return nil
	}
	if payload.UserID == 0 || payload.AvatarURL == "" {
		helper.LogError(logger, helper.OpMQConsumer, "invalid AvatarModeration payload", nil, map[string]interface{}{
			"user_id":    payload.UserID,
			"avatar_url": payload.AvatarURL,
			"request_id": payload.RequestID,
		})
		return nil
	}

	helper.LogInfo(logger, helper.OpMQConsumer, "processing avatar moderation", map[string]interface{}{
		"user_id":    payload.UserID,
		"avatar_url": payload.AvatarURL,
		"request_id": payload.RequestID,
	})

	if payload.DefaultAvatarURL != "" {
		helper.LogInfo(logger, helper.OpMQConsumer, "setting default avatar", map[string]interface{}{
			"user_id":        payload.UserID,
			"default_avatar": payload.DefaultAvatarURL,
		})
		if err := updateUserAvatar(ctx, svcCtx, payload.UserID, payload.DefaultAvatarURL); err != nil {
			helper.LogError(logger, helper.OpMQConsumer, "update default avatar failed", err, map[string]interface{}{
				"user_id": payload.UserID,
			})
			return err
		}
	}

	if svcCtx.AgentRPC == nil {
		helper.LogError(logger, helper.OpMQConsumer, "agent rpc not initialized", nil, map[string]interface{}{
			"user_id": payload.UserID,
		})
		return status.Error(codes.Unavailable, "agent rpc not initialized")
	}

	helper.LogInfo(logger, helper.OpMQConsumer, "calling agent moderation service", map[string]interface{}{
		"user_id": payload.UserID,
	})
	moderateCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	resp, err := svcCtx.AgentRPC.ModerateAvatar(moderateCtx, &agentclient.ModerateAvatarRequest{
		UserId:    payload.UserID,
		ImageUrl:  payload.AvatarURL,
		Scene:     "avatar",
		RequestId: payload.RequestID,
	})
	if err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "call moderation service failed", err, map[string]interface{}{
			"user_id":    payload.UserID,
			"request_id": payload.RequestID,
		})
		return err
	}

	helper.LogInfo(logger, helper.OpMQConsumer, "moderation service response", map[string]interface{}{
		"user_id":    payload.UserID,
		"decision":   resp.GetDecision().String(),
		"request_id": payload.RequestID,
	})

	if resp == nil || resp.GetDecision() != agent.ModerationDecision_PASS {
		helper.LogInfo(logger, helper.OpMQConsumer, "avatar moderation rejected", map[string]interface{}{
			"user_id":  payload.UserID,
			"decision": resp.GetDecision().String(),
		})
		return nil
	}

	helper.LogInfo(logger, helper.OpMQConsumer, "avatar moderation passed, updating avatar", map[string]interface{}{
		"user_id":    payload.UserID,
		"avatar_url": payload.AvatarURL,
	})
	if err := updateUserAvatar(ctx, svcCtx, payload.UserID, payload.AvatarURL); err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "update approved avatar failed", err, map[string]interface{}{
			"user_id": payload.UserID,
		})
		return err
	}

	helper.LogSuccess(logger, helper.OpMQConsumer, map[string]interface{}{
		"event":      "avatar_moderation_passed",
		"user_id":    payload.UserID,
		"avatar_url": payload.AvatarURL,
	})
	return nil
}

func updateUserAvatar(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64, avatarURL string) error {
	logger := logx.WithContext(ctx)

	if userID == 0 || avatarURL == "" {
		return nil
	}

	db := svcCtx.DB().WithContext(ctx)
	result := db.Model(&model.User{}).Where("id = ?", userID).Update("avatar_url", avatarURL)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		helper.LogError(logger, helper.OpMQConsumer, "user not found when updating avatar", nil, map[string]interface{}{
			"user_id": userID,
		})
		return nil
	}
	return nil
}
