package logic

import (
	"context"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (l *UnfollowUserLogic) UnfollowUser(in *user.UnfollowUserRequest) (*user.UnfollowUserResponse, error) {
	// 1. 验证参数
	if in.OperatorId == 0 {
		return nil, status.Error(codes.InvalidArgument, "operator_id is required")
	}
	if in.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// 2. 防止自我操作
	if in.OperatorId == in.UserId {
		return nil, status.Error(codes.InvalidArgument, "cannot unfollow yourself")
	}

	// 3. 检查关注关系是否存在
	result := l.svcCtx.DB().Where("follower_id = ? AND following_id = ?", in.OperatorId, in.UserId).Delete(&model.FollowRelation{})
	if result.Error != nil {
		l.Errorf("delete follow relation failed: %v", result.Error)
		return nil, status.Error(codes.Internal, "unfollow failed")
	}

	if result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "you haven't followed this user")
	}

	l.Infof("user %d unfollowed user %d", in.OperatorId, in.UserId)

	return &user.UnfollowUserResponse{
		Success: true,
		Message: "unfollow user success",
	}, nil
}
