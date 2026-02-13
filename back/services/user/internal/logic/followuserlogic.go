package logic

import (
	"context"
	"encoding/json"
	"errors"

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
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
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
		return nil, status.Error(codes.InvalidArgument, "cannot follow yourself")
	}

	// 3. 验证被关注用户是否存在
	var targetUser model.User
	if err := l.svcCtx.DB().Where("id = ?", in.UserId).First(&targetUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "target user not found")
		}
		l.Errorf("get target user failed: %v", err)
		return nil, status.Error(codes.Internal, "get target user failed")
	}

	// 4. 检查是否已经关注
	var existing model.FollowRelation
	if err := l.svcCtx.DB().Where("follower_id = ? AND following_id = ?", in.OperatorId, in.UserId).First(&existing).Error; err == nil {
		return nil, status.Error(codes.AlreadyExists, "you have already followed this user")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Errorf("check follow status failed: %v", err)
		return nil, status.Error(codes.Internal, "check follow status failed")
	}

	// 5. 检查关注人数上限（最多1000人）
	const maxFollowingCount = 1000
	var followingCount int64

	// 优先从Redis获取关注数
	if l.svcCtx.UserCache != nil {
		count, err := l.svcCtx.UserCache.GetFollowingCount(int64(in.OperatorId))
		if err == nil {
			// 缓存命中
			followingCount = count
		} else {
			// Redis未命中，从数据库查询
			if err := l.svcCtx.DB().Model(&model.FollowRelation{}).Where("follower_id = ?", in.OperatorId).Count(&followingCount).Error; err != nil {
				l.Errorf("get following count from db failed: %v", err)
				return nil, status.Error(codes.Internal, "get following count failed")
			}
			// 回写Redis
			l.svcCtx.UserCache.SetFollowingCount(int64(in.OperatorId), followingCount)
		}
	} else {
		// Redis不可用，从数据库查询
		if err := l.svcCtx.DB().Model(&model.FollowRelation{}).Where("follower_id = ?", in.OperatorId).Count(&followingCount).Error; err != nil {
			l.Errorf("get following count from db failed: %v", err)
			return nil, status.Error(codes.Internal, "get following count failed")
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
				l.Errorf("send follow user event failed: %v", err)
			}
		}
	}

	l.Infof("user %d followed user %d", in.OperatorId, in.UserId)

	return &user.FollowUserResponse{
		Success: true,
		Message: "follow user success",
	}, nil
}
