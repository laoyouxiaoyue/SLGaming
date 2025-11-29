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

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	if l.svcCtx.UserRPC == nil {
		return nil, fmt.Errorf("user rpc client not initialized")
	}

	// 调用用户服务的 RPC
	rpcResp, err := l.svcCtx.UserRPC.Login(l.ctx, &userclient.LoginRequest{
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		l.Errorf("call user rpc failed: %v", err)
		return &types.LoginResponse{
			BaseResp: types.BaseResp{
				Code: 500,
				Msg:  "登录失败: " + err.Error(),
			},
		}, nil
	}

	// 生成 JWT token
	accessToken, err := l.svcCtx.JWT.GenerateToken(rpcResp.Id)
	if err != nil {
		l.Errorf("generate jwt token failed: %v", err)
		return &types.LoginResponse{
			BaseResp: types.BaseResp{
				Code: 500,
				Msg:  "生成 token 失败: " + err.Error(),
			},
		}, nil
	}

	return &types.LoginResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: types.LoginData{
			AccessToken: accessToken,
		},
	}, nil
}
