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
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.LoginByCodeResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	// 验证验证码
	if l.svcCtx.CodeRPC != nil {
		verifyResp, err := l.svcCtx.CodeRPC.VerifyCode(l.ctx, &codeclient.VerifyCodeRequest{
			Phone:   req.Phone,
			Purpose: "login",
			Code:    req.Code,
		})
		if err != nil {
			code, msg := utils.HandleRPCError(err, l.Logger, "VerifyCode")
			return &types.LoginByCodeResponse{
				BaseResp: types.BaseResp{
					Code: code,
					Msg:  msg,
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
		code, msg := utils.HandleRPCError(err, l.Logger, "LoginByCode")
		return &types.LoginByCodeResponse{
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
		return &types.LoginByCodeResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	return &types.LoginByCodeResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: *tokenData,
	}, nil
}
