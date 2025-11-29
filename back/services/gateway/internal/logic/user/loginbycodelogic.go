// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"fmt"

	"SLGaming/back/services/code/codeclient"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginByCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginByCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByCodeLogic {
	return &LoginByCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginByCodeLogic) LoginByCode(req *types.LoginByCodeRequest) (resp *types.LoginByCodeResponse, err error) {
	if l.svcCtx.UserRPC == nil {
		return nil, fmt.Errorf("user rpc client not initialized")
	}

	// 验证验证码
	if l.svcCtx.CodeRPC != nil {
		verifyResp, err := l.svcCtx.CodeRPC.VerifyCode(l.ctx, &codeclient.VerifyCodeRequest{
			Phone:   req.Phone,
			Purpose: "login",
			Code:    req.Code,
		})
		if err != nil {
			l.Errorf("verify code failed: %v", err)
			return &types.LoginByCodeResponse{
				BaseResp: types.BaseResp{
					Code: 400,
					Msg:  "验证码验证失败: " + err.Error(),
				},
			}, nil
		}
		if !verifyResp.Passed {
			return &types.LoginByCodeResponse{
				BaseResp: types.BaseResp{
					Code: 400,
					Msg:  "验证码错误或已过期",
				},
			}, nil
		}
	}

	// 调用用户服务的 RPC
	rpcResp, err := l.svcCtx.UserRPC.LoginByCode(l.ctx, &userclient.LoginByCodeRequest{
		Phone: req.Phone,
	})
	if err != nil {
		l.Errorf("call user rpc failed: %v", err)
		return &types.LoginByCodeResponse{
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
		return &types.LoginByCodeResponse{
			BaseResp: types.BaseResp{
				Code: 500,
				Msg:  "生成 token 失败: " + err.Error(),
			},
		}, nil
	}

	return &types.LoginByCodeResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: types.LoginData{
			AccessToken: accessToken,
		},
	}, nil
}
