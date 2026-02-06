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
		logx.Infof("avatar moderation consumer not started: rocketmq not configured")
		return
	}

	consumer, err := initAvatarConsumer(cfg, svcCtx)
	if err != nil {
		logx.Errorf("init avatar moderation consumer failed: %v", err)
		return
	}

	go func() {
		<-ctx.Done()
		shutdownAvatarConsumer(consumer)
	}()

	logx.Infof("avatar moderation consumer started, topic=%s", avatarmq.AvatarEventTopic())
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
	if msg.GetTags() != avatarmq.EventTypeAvatarSubmit() {
		return nil
	}

	var payload avatarmq.AvatarModerationPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logx.Errorf("unmarshal avatar moderation payload failed: %v, body=%s", err, string(msg.Body))
		return nil
	}
	if payload.UserID == 0 || payload.AvatarURL == "" {
		logx.Errorf("invalid avatar moderation payload: user_id=%d, avatar_url=%s", payload.UserID, payload.AvatarURL)
		return nil
	}

	if payload.DefaultAvatarURL != "" {
		if err := updateUserAvatar(ctx, svcCtx, payload.UserID, payload.DefaultAvatarURL); err != nil {
			return err
		}
	}

	if svcCtx.AgentRPC == nil {
		return status.Error(codes.Unavailable, "agent rpc not initialized")
	}

	moderateCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	resp, err := svcCtx.AgentRPC.ModerateAvatar(moderateCtx, &agentclient.ModerateAvatarRequest{
		UserId:    payload.UserID,
		ImageUrl:  payload.AvatarURL,
		Scene:     "avatar",
		RequestId: payload.RequestID,
	})
	if err != nil {
		return err
	}
	if resp == nil || resp.GetDecision() != agent.ModerationDecision_PASS {
		return nil
	}

	return updateUserAvatar(ctx, svcCtx, payload.UserID, payload.AvatarURL)
}

func updateUserAvatar(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64, avatarURL string) error {
	if userID == 0 || avatarURL == "" {
		return nil
	}

	db := svcCtx.DB().WithContext(ctx)
	result := db.Model(&model.User{}).Where("id = ?", userID).Update("avatar_url", avatarURL)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		logx.Errorf("user not found when updating avatar: user_id=%d", userID)
		return nil
	}
	return nil
}
