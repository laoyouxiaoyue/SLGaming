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

type FollowUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFollowUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowUserLogic {
	return &FollowUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// fallbackUpdateFollowCounts 降级处理：直接同步更新数据库中的粉丝数和关注数
func (l *FollowUserLogic) fallbackUpdateFollowCounts(followerID, followingID uint64) error {
	return l.svcCtx.DB().Transaction(func(tx *gorm.DB) error {
		// 增加被关注用户的粉丝数
		if err := tx.Model(&model.User{}).Where("id = ?", followingID).
			UpdateColumn("follower_count", gorm.Expr("follower_count + ?", 1)).Error; err != nil {
			return err
		}

		// 增加关注者的关注数
		if err := tx.Model(&model.User{}).Where("id = ?", followerID).
			UpdateColumn("following_count", gorm.Expr("following_count + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})
}

func (l *FollowUserLogic) FollowUser(in *user.FollowUserRequest) (*user.FollowUserResponse, error) {
	// 1. 验证参数
	if in.OperatorId == 0 {
		return nil, status.Error(codes.InvalidArgument, "operator_id is required")
	}
	if in.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// 2. 防止自我关注
	if in.OperatorId == in.UserId {
		metrics.FollowTotal.WithLabelValues("error", "follow").Inc()
		return nil, status.Error(codes.InvalidArgument, "cannot follow yourself")
	}

	// 3. 布隆过滤器快速检查被关注用户是否存在
	// 如果布隆过滤器说"不存在"，那用户一定不存在，直接返回（省去数据库查询）
	if l.svcCtx.BloomFilter != nil {
		exists, err := l.svcCtx.BloomFilter.UserID.MightContain(l.ctx, int64(in.UserId))
		if err != nil {
			l.Errorf("bloom filter check user failed: %v", err)
			// 布隆过滤器查询失败，降级到数据库查询
		} else if !exists {
			// 用户一定不存在，直接返回
			metrics.FollowTotal.WithLabelValues("error", "follow").Inc()
			return nil, status.Error(codes.NotFound, "target user not found")
		}
		// 如果存在，需要查数据库确认（布隆过滤器有假阳性）
	}

	// 4. 合并查询：验证用户是否存在 + 是否已关注（减少数据库调用）
	var result struct {
		UserExists bool
		Followed   bool
	}
	if err := l.svcCtx.DB().Raw(`
		SELECT 
			EXISTS(SELECT 1 FROM users WHERE id = ?) as user_exists,
			EXISTS(SELECT 1 FROM follow_relations WHERE follower_id = ? AND following_id = ?) as followed
	`, in.UserId, in.OperatorId, in.UserId).Scan(&result).Error; err != nil {
		l.Errorf("check user and follow status failed: %v", err)
		metrics.FollowTotal.WithLabelValues("error", "follow").Inc()
		return nil, status.Error(codes.Internal, "check follow status failed")
	}

	if !result.UserExists {
		metrics.FollowTotal.WithLabelValues("error", "follow").Inc()
		return nil, status.Error(codes.NotFound, "target user not found")
	}

	if result.Followed {
		metrics.FollowTotal.WithLabelValues("duplicate", "follow").Inc()
		return nil, status.Error(codes.AlreadyExists, "you have already followed this user")
	}

	// 5. 检查关注人数上限（最多1000人）
	const maxFollowingCount = 1000
	var followingCount int64

	// 优先从 Redis 获取关注数
	if l.svcCtx.UserCache != nil {
		count, err := l.svcCtx.UserCache.GetFollowingCount(int64(in.OperatorId))
		if err == nil {
			followingCount = count
		}
	}

	// Redis 未命中或不可用，从用户表获取 following_count（比 COUNT 快）
	if followingCount == 0 {
		var user model.User
		if err := l.svcCtx.DB().Select("following_count").Where("id = ?", in.OperatorId).First(&user).Error; err != nil {
			l.Errorf("get user following count failed: %v", err)
			return nil, status.Error(codes.Internal, "get following count failed")
		}
		followingCount = user.FollowingCount
		// 回写 Redis
		if l.svcCtx.UserCache != nil {
			l.svcCtx.UserCache.SetFollowingCount(int64(in.OperatorId), followingCount)
		}
	}

	if followingCount >= maxFollowingCount {
		return nil, status.Error(codes.ResourceExhausted, "you can only follow up to 1000 users")
	}

	// 6. 创建关注关系
	followRelation := model.FollowRelation{
		FollowerID:  in.OperatorId,
		FollowingID: in.UserId,
	}

	if err := l.svcCtx.DB().Create(&followRelation).Error; err != nil {
		l.Errorf("create follow relation failed: %v", err)
		metrics.FollowTotal.WithLabelValues("error", "follow").Inc()
		return nil, status.Error(codes.Internal, "create follow relation failed")
	}

	// 7. 更新Redis缓存中的粉丝数和关注数
	if l.svcCtx.UserCache != nil {
		// 增加被关注用户的粉丝数
		if err := l.svcCtx.UserCache.IncrFollowerCount(int64(in.UserId)); err != nil {
			l.Errorf("incr follower count cache failed: %v", err)
		}
		// 增加关注者的关注数
		if err := l.svcCtx.UserCache.IncrFollowingCount(int64(in.OperatorId)); err != nil {
			l.Errorf("incr following count cache failed: %v", err)
		}
	}

	// 8. 发送关注事件到消息队列，异步更新数据库计数
	// 降级策略：如果MQ发送失败，直接同步更新数据库计数
	mqSendSuccess := false
	if l.svcCtx.EventProducer != nil {
		payload := userMQ.FollowUserPayload{
			FollowerID:  in.OperatorId,
			FollowingID: in.UserId,
		}
		payloadJSON, err := json.Marshal(payload)
		if err == nil {
			msg := primitive.NewMessage(userMQ.FollowEventTopic(), payloadJSON)
			msg.WithTag(userMQ.EventTypeFollowUser())
			_, err := l.svcCtx.EventProducer.SendSync(l.ctx, msg)
			if err != nil {
				l.Errorf("send follow user event failed: %v, will fallback to direct db update", err)
			} else {
				mqSendSuccess = true
			}
		}
	}

	// 降级处理：MQ发送失败时，直接同步更新数据库计数
	if !mqSendSuccess {
		l.Infof("mq send failed or producer not available, fallback to direct db update for follow: follower_id=%d, following_id=%d", in.OperatorId, in.UserId)
		if err := l.fallbackUpdateFollowCounts(in.OperatorId, in.UserId); err != nil {
			l.Errorf("fallback update follow counts failed: %v", err)
			// 降级失败不影响主流程，只是记录错误
		}
	}

	l.Infof("user %d followed user %d", in.OperatorId, in.UserId)

	metrics.FollowTotal.WithLabelValues("success", "follow").Inc()

	return &user.FollowUserResponse{
		Success: true,
		Message: "follow user success",
	}, nil
}
