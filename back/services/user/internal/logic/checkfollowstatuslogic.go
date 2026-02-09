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

type CheckFollowStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckFollowStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckFollowStatusLogic {
	return &CheckFollowStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckFollowStatusLogic) CheckFollowStatus(in *user.CheckFollowStatusRequest) (*user.CheckFollowStatusResponse, error) {
	// 1. 验证参数
	if in.OperatorId == 0 {
		return nil, status.Error(codes.InvalidArgument, "operator_id is required")
	}
	if in.TargetUserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "target_user_id is required")
	}

	// 2. 自我检查
	if in.OperatorId == in.TargetUserId {
		return &user.CheckFollowStatusResponse{
			IsFollowing: false,
			IsFollowed:  false,
			IsMutual:    false,
		}, nil
	}

	// 3. 检查当前用户是否关注目标用户
	var isFollowing bool
	if err := l.svcCtx.DB().Model(&model.FollowRelation{}).
		Where("follower_id = ? AND following_id = ?", in.OperatorId, in.TargetUserId).
		Select("COUNT(*) > 0").
		Scan(&isFollowing).Error; err != nil {
		l.Errorf("check following status failed: %v", err)
		return nil, status.Error(codes.Internal, "check following status failed")
	}

	// 4. 检查目标用户是否关注当前用户
	var isFollowed bool
	if err := l.svcCtx.DB().Model(&model.FollowRelation{}).
		Where("follower_id = ? AND following_id = ?", in.TargetUserId, in.OperatorId).
		Select("COUNT(*) > 0").
		Scan(&isFollowed).Error; err != nil {
		l.Errorf("check followed status failed: %v", err)
		return nil, status.Error(codes.Internal, "check followed status failed")
	}

	// 5. 计算是否互相关注
	isMutual := isFollowing && isFollowed

	l.Infof("check follow status: operator=%d, target=%d, is_following=%v, is_followed=%v, is_mutual=%v",
		in.OperatorId, in.TargetUserId, isFollowing, isFollowed, isMutual)

	return &user.CheckFollowStatusResponse{
		IsFollowing: isFollowing,
		IsFollowed:  isFollowed,
		IsMutual:    isMutual,
	}, nil
}
