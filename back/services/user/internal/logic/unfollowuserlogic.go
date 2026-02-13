package logic

import (
	"context"
	"encoding/json"

	"SLGaming/back/services/user/internal/metrics"
	"SLGaming/back/services/user/internal/model"
	userMQ "SLGaming/back/services/user/internal/mq"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type UnfollowUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnfollowUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfollowUserLogic {
	return &UnfollowUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// fallbackUpdateUnfollowCounts 降级处理：直接同步更新数据库中的粉丝数和关注数（取消关注）
func (l *UnfollowUserLogic) fallbackUpdateUnfollowCounts(followerID, followingID uint64) error {
	return l.svcCtx.DB().Transaction(func(tx *gorm.DB) error {
		// 减少被关注用户的粉丝数（确保不会小于0）
		if err := tx.Model(&model.User{}).Where("id = ?", followingID).
			UpdateColumn("follower_count", gorm.Expr("GREATEST(follower_count - ?, 0)", 1)).Error; err != nil {
			return err
		}

		// 减少关注者的关注数（确保不会小于0）
		if err := tx.Model(&model.User{}).Where("id = ?", followerID).
			UpdateColumn("following_count", gorm.Expr("GREATEST(following_count - ?, 0)", 1)).Error; err != nil {
			return err
		}

		return nil
	})
}

func (l *UnfollowUserLogic) UnfollowUser(in *user.UnfollowUserRequest) (*user.UnfollowUserResponse, error) {
	// 1. 验证参数
	if in.OperatorId == 0 {
		metrics.FollowTotal.WithLabelValues("error", "unfollow").Inc()
		return nil, status.Error(codes.InvalidArgument, "operator_id is required")
	}
	if in.UserId == 0 {
		metrics.FollowTotal.WithLabelValues("error", "unfollow").Inc()
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// 2. 防止自我操作
	if in.OperatorId == in.UserId {
		metrics.FollowTotal.WithLabelValues("error", "unfollow").Inc()
		return nil, status.Error(codes.InvalidArgument, "cannot unfollow yourself")
	}

	// 3. 检查关注关系是否存在
	result := l.svcCtx.DB().Where("follower_id = ? AND following_id = ?", in.OperatorId, in.UserId).Delete(&model.FollowRelation{})
	if result.Error != nil {
		l.Errorf("delete follow relation failed: %v", result.Error)
		metrics.FollowTotal.WithLabelValues("error", "unfollow").Inc()
		return nil, status.Error(codes.Internal, "unfollow failed")
	}

	if result.RowsAffected == 0 {
		metrics.FollowTotal.WithLabelValues("not_found", "unfollow").Inc()
		return nil, status.Error(codes.NotFound, "you haven't followed this user")
	}

	// 更新Redis缓存中的粉丝数和关注数
	if l.svcCtx.UserCache != nil {
		// 减少被关注用户的粉丝数
		if err := l.svcCtx.UserCache.DecrFollowerCount(int64(in.UserId)); err != nil {
			l.Errorf("decr follower count cache failed: %v", err)
		}
		// 减少关注者的关注数
		if err := l.svcCtx.UserCache.DecrFollowingCount(int64(in.OperatorId)); err != nil {
			l.Errorf("decr following count cache failed: %v", err)
		}
	}

	// 发送取消关注事件到消息队列，异步更新数据库计数
	// 降级策略：如果MQ发送失败，直接同步更新数据库计数
	mqSendSuccess := false
	if l.svcCtx.EventProducer != nil {
		payload := userMQ.UnfollowUserPayload{
			FollowerID:  in.OperatorId,
			FollowingID: in.UserId,
		}
		payloadJSON, err := json.Marshal(payload)
		if err == nil {
			msg := primitive.NewMessage(userMQ.FollowEventTopic(), payloadJSON)
			msg.WithTag(userMQ.EventTypeUnfollowUser())
			_, err := l.svcCtx.EventProducer.SendSync(l.ctx, msg)
			if err != nil {
				l.Errorf("send unfollow user event failed: %v, will fallback to direct db update", err)
			} else {
				mqSendSuccess = true
			}
		}
	}

	// 降级处理：MQ发送失败时，直接同步更新数据库计数
	if !mqSendSuccess {
		l.Infof("mq send failed or producer not available, fallback to direct db update for unfollow: follower_id=%d, following_id=%d", in.OperatorId, in.UserId)
		if err := l.fallbackUpdateUnfollowCounts(in.OperatorId, in.UserId); err != nil {
			l.Errorf("fallback update unfollow counts failed: %v", err)
			// 降级失败不影响主流程，只是记录错误
		}
	}

	l.Infof("user %d unfollowed user %d", in.OperatorId, in.UserId)

	metrics.FollowTotal.WithLabelValues("success", "unfollow").Inc()

	return &user.UnfollowUserResponse{
		Success: true,
		Message: "unfollow user success",
	}, nil
}
