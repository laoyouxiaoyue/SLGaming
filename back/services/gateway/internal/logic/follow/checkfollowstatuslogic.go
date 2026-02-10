// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package follow

import (
	"context"

	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckFollowStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckFollowStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckFollowStatusLogic {
	return &CheckFollowStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckFollowStatusLogic) CheckFollowStatus(req *types.CheckFollowStatusRequest) (resp *types.CheckFollowStatusResponse, err error) {
	if l.svcCtx.UserRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.CheckFollowStatusResponse{BaseResp: types.BaseResp{Code: code, Msg: msg}}, nil
	}

	userID, err := middleware.GetUserID(l.ctx)
	if err != nil || userID == 0 {
		return &types.CheckFollowStatusResponse{
			BaseResp: types.BaseResp{Code: 401, Msg: "未登录或认证失败"},
		}, nil
	}

	rpcResp, err := l.svcCtx.UserRPC.CheckFollowStatus(l.ctx, &userclient.CheckFollowStatusRequest{
		OperatorId:   userID,
		TargetUserId: req.TargetUserId,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "CheckFollowStatus")
		return &types.CheckFollowStatusResponse{BaseResp: types.BaseResp{Code: code, Msg: msg}}, nil
	}

	return &types.CheckFollowStatusResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		Data: types.CheckFollowStatusData{
			IsFollowing: rpcResp.GetIsFollowing(),
			IsFollowed:  rpcResp.GetIsFollowed(),
			IsMutual:    rpcResp.GetIsMutual(),
		},
	}, nil
}
