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

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	if l.svcCtx.UserRPC == nil {
		return nil, fmt.Errorf("user rpc client not initialized")
	}

	// 验证验证码
	if l.svcCtx.CodeRPC != nil {
		verifyResp, err := l.svcCtx.CodeRPC.VerifyCode(l.ctx, &codeclient.VerifyCodeRequest{
			Phone:   req.Phone,
			Purpose: "register",
			Code:    req.Code,
		})
		if err != nil {
			l.Errorf("verify code failed: %v", err)
			return &types.RegisterResponse{
				BaseResp: types.BaseResp{
					Code: 400,
					Msg:  "验证码验证失败: " + err.Error(),
				},
			}, nil
		}
		if !verifyResp.Passed {
			return &types.RegisterResponse{
				BaseResp: types.BaseResp{
					Code: 400,
					Msg:  "验证码错误或已过期",
				},
			}, nil
		}
	}

	// 调用用户服务的 RPC
	rpcResp, err := l.svcCtx.UserRPC.Register(l.ctx, &userclient.RegisterRequest{
		Phone:    req.Phone,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		l.Errorf("call user rpc failed: %v", err)
		return &types.RegisterResponse{
			BaseResp: types.BaseResp{
				Code: 500,
				Msg:  "注册失败: " + err.Error(),
			},
		}, nil
	}

	// 生成 Access Token 和 Refresh Token
	tokenData, err := generateTokens(l.ctx, l.svcCtx, rpcResp.Id, l.Logger)
	if err != nil {
		return &types.RegisterResponse{
			BaseResp: types.BaseResp{
				Code: 500,
				Msg:  "生成 token 失败: " + err.Error(),
			},
		}, nil
	}

	return &types.RegisterResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: types.RegisterData{
			AccessToken:  tokenData.AccessToken,
			RefreshToken: tokenData.RefreshToken,
			ExpiresIn:    tokenData.ExpiresIn,
		},
	}, nil
}
