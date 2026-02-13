// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"SLGaming/back/services/code/codeclient"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
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
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.RegisterResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	// 验证验证码
	if l.svcCtx.CodeRPC != nil {
		verifyResp, err := l.svcCtx.CodeRPC.VerifyCode(l.ctx, &codeclient.VerifyCodeRequest{
			Phone:   req.Phone,
			Purpose: "register",
			Code:    req.Code,
		})
		if err != nil {
			code, msg := utils.HandleRPCError(err, l.Logger, "VerifyCode")
			return &types.RegisterResponse{
				BaseResp: types.BaseResp{
					Code: code,
					Msg:  msg,
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
		Role:     int32(req.Role),
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "Register")
		return &types.RegisterResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 生成 Access Token 和 Refresh Token
	tokenData, err := generateTokens(l.ctx, l.svcCtx, rpcResp.Id, l.Logger)
	if err != nil {
		code, msg := utils.HandleError(err, l.Logger, "GenerateTokens")
		return &types.RegisterResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	return &types.RegisterResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  utils.GetSuccessMsg("Register"),
		},
		Data: types.RegisterData{
			AccessToken:  tokenData.AccessToken,
			RefreshToken: tokenData.RefreshToken,
			ExpiresIn:    tokenData.ExpiresIn,
		},
	}, nil
}
