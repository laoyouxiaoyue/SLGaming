// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"fmt"

	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserLogic) UpdateUser(req *types.UpdateUserRequest) (resp *types.UpdateUserResponse, err error) {
	if l.svcCtx.UserRPC == nil {
		return nil, fmt.Errorf("user rpc client not initialized")
	}

	// 调用用户服务的 RPC
	rpcResp, err := l.svcCtx.UserRPC.UpdateUser(l.ctx, &userclient.UpdateUserRequest{
		Id:       req.Id,
		Nickname: req.Nickname,
		Password: req.Password,
		Phone:    req.Phone,
	})
	if err != nil {
		l.Errorf("call user rpc failed: %v", err)
		return &types.UpdateUserResponse{
			BaseResp: types.BaseResp{
				Code: 500,
				Msg:  "更新用户信息失败: " + err.Error(),
			},
		}, nil
	}

	return &types.UpdateUserResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: types.UserInfo{
			Id:       rpcResp.User.Id,
			Uid:      rpcResp.User.Uid,
			Nickname: rpcResp.User.Nickname,
			Phone:    rpcResp.User.Phone,
		},
	}, nil
}
