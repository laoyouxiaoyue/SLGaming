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

type GetUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserLogic) GetUser(req *types.GetUserRequest) (*types.GetUserResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	if req.Id == 0 && req.Uid == 0 && req.Phone == "" {
		return nil, errors.New("must provide id, uid or phone")
	}

	rpc, err := getRPC("user", l.svcCtx)
	if err != nil {
		return nil, err
	}
	userRPC := rpc.(userclient.User)

	rpcResp, err := userRPC.GetUser(l.ctx, &userclient.GetUserRequest{
		Id:    req.Id,
		Uid:   req.Uid,
		Phone: req.Phone,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetUserResponse{
		BaseResp: successResp(),
		Data:     toUserInfo(rpcResp.GetUser()),
	}, nil
}
