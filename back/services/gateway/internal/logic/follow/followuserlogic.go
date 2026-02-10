// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package follow

import (
	"context"

	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type FollowUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFollowUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowUserLogic {
	return &FollowUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FollowUserLogic) FollowUser(req *types.FollowUserRequest) (resp *types.FollowUserResponse, err error) {
	if l.svcCtx.UserRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.FollowUserResponse{BaseResp: types.BaseResp{Code: code, Msg: msg}}, nil
	}

	rpcResp, err := l.svcCtx.UserRPC.FollowUser(l.ctx, &userclient.FollowUserRequest{
		OperatorId: req.OperatorId,
		UserId:     req.UserId,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "FollowUser")
		return &types.FollowUserResponse{BaseResp: types.BaseResp{Code: code, Msg: msg}}, nil
	}

	return &types.FollowUserResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		Data:     types.FollowUserData{Success: rpcResp.Success, Message: rpcResp.Message},
	}, nil
}
