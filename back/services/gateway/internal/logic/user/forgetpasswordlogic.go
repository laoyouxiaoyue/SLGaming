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

type ForgetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewForgetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ForgetPasswordLogic {
	return &ForgetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ForgetPasswordLogic) ForgetPassword(req *types.ForgetPasswordRequest) (*types.ForgetPasswordResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}

	if err := verifyCode(l.ctx, l.svcCtx, req.Phone, req.Code, "forget_password"); err != nil {
		return nil, err
	}

	rpc, err := getRPC("user", l.svcCtx)
	if err != nil {
		return nil, err
	}
	userRPC := rpc.(userclient.User)

	rpcResp, err := userRPC.ForgetPassword(l.ctx, &userclient.ForgetPasswordRequest{
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	token, err := generateAccessToken(l.ctx, l.svcCtx, rpcResp.GetId())
	if err != nil {
		return nil, err
	}

	return &types.ForgetPasswordResponse{
		BaseResp: successResp(),
		Data: types.LoginData{
			AccessToken: token,
		},
	}, nil
}
