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

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (*types.RegisterResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	if req.Phone == "" {
		return nil, errors.New("phone is required")
	}

	if err := verifyCode(l.ctx, l.svcCtx, req.Phone, req.Code, "register"); err != nil {
		return nil, err
	}

	rpc, err := getRPC("user", l.svcCtx)
	if err != nil {
		return nil, err
	}
	userRPC := rpc.(userclient.User)

	if _, err := userRPC.Register(l.ctx, &userclient.RegisterRequest{
		Phone:    req.Phone,
		Password: req.Password,
		Nickname: req.Nickname,
	}); err != nil {
		return nil, err
	}

	loginResp, err := userRPC.Login(l.ctx, &userclient.LoginRequest{
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	token, err := generateAccessToken(l.ctx, l.svcCtx, loginResp.GetId())
	if err != nil {
		return nil, err
	}

	return &types.RegisterResponse{
		BaseResp: successResp(),
		Data: types.RegisterData{
			AccessToken: token,
		},
	}, nil
}
