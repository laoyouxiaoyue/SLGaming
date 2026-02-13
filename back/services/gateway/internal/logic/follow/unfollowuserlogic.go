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

type UnfollowUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnfollowUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfollowUserLogic {
	return &UnfollowUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnfollowUserLogic) UnfollowUser(req *types.UnfollowUserRequest) (resp *types.UnfollowUserResponse, err error) {
	if l.svcCtx.UserRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.UnfollowUserResponse{BaseResp: types.BaseResp{Code: code, Msg: msg}}, nil
	}

	userID, err := middleware.GetUserID(l.ctx)
	if err != nil || userID == 0 {
		return &types.UnfollowUserResponse{
			BaseResp: types.BaseResp{Code: 401, Msg: "未登录或认证失败"},
		}, nil
	}

	rpcResp, err := l.svcCtx.UserRPC.UnfollowUser(l.ctx, &userclient.UnfollowUserRequest{
		OperatorId: userID,
		UserId:     req.UserId,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "UnfollowUser")
		return &types.UnfollowUserResponse{BaseResp: types.BaseResp{Code: code, Msg: msg}}, nil
	}

	return &types.UnfollowUserResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		Data:     types.UnfollowUserData{Success: rpcResp.Success, Message: rpcResp.Message},
	}, nil
}
