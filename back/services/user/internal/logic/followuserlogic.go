package logic

import (
	"context"
	"errors"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

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

	// 5. 创建关注关系
	followRelation := model.FollowRelation{
		FollowerID:  in.OperatorId,
		FollowingID: in.UserId,
	}

	if err := l.svcCtx.DB().Create(&followRelation).Error; err != nil {
		l.Errorf("create follow relation failed: %v", err)
		return nil, status.Error(codes.Internal, "create follow relation failed")
	}

	l.Infof("user %d followed user %d", in.OperatorId, in.UserId)

	return &user.FollowUserResponse{
		Success: true,
		Message: "follow user success",
	}, nil
}
