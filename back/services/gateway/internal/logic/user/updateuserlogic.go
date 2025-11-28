// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"errors"

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

func (l *UpdateUserLogic) UpdateUser(req *types.UpdateUserRequest) (*types.UpdateUserResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	if req.Id == 0 {
		return nil, errors.New("id is required")
	}

	rpc, err := getRPC("user", l.svcCtx)
	if err != nil {
		return nil, err
	}
	userRPC := rpc.(userclient.User)

	rpcResp, err := userRPC.UpdateUser(l.ctx, &userclient.UpdateUserRequest{
		Id:       req.Id,
		Nickname: req.Nickname,
		Password: req.Password,
		Phone:    req.Phone,
	})
	if err != nil {
		return nil, err
	}

	return &types.UpdateUserResponse{
		BaseResp: successResp(),
		Data:     toUserInfo(rpcResp.GetUser()),
	}, nil
}
