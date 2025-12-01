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

func (l *ForgetPasswordLogic) ForgetPassword(req *types.ForgetPasswordRequest) (resp *types.ForgetPasswordResponse, err error) {
	if l.svcCtx.UserRPC == nil {
		return nil, fmt.Errorf("user rpc client not initialized")
	}

	// 验证验证码
	if l.svcCtx.CodeRPC != nil {
		verifyResp, err := l.svcCtx.CodeRPC.VerifyCode(l.ctx, &codeclient.VerifyCodeRequest{
			Phone:   req.Phone,
			Purpose: "forget_password",
			Code:    req.Code,
		})
		if err != nil {
			l.Errorf("verify code failed: %v", err)
			return &types.ForgetPasswordResponse{
				BaseResp: types.BaseResp{
					Code: 400,
					Msg:  "验证码验证失败: " + err.Error(),
				},
			}, nil
		}
		if !verifyResp.Passed {
			return &types.ForgetPasswordResponse{
				BaseResp: types.BaseResp{
					Code: 400,
					Msg:  "验证码错误或已过期",
				},
			}, nil
		}
	}

	// 调用用户服务的 RPC
	rpcResp, err := l.svcCtx.UserRPC.ForgetPassword(l.ctx, &userclient.ForgetPasswordRequest{
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		l.Errorf("call user rpc failed: %v", err)
		return &types.ForgetPasswordResponse{
			BaseResp: types.BaseResp{
				Code: 500,
				Msg:  "重置密码失败: " + err.Error(),
			},
		}, nil
	}

	// 生成 Access Token 和 Refresh Token
	tokenData, err := generateTokens(l.ctx, l.svcCtx, rpcResp.Id, l.Logger)
	if err != nil {
		return &types.ForgetPasswordResponse{
			BaseResp: types.BaseResp{
				Code: 500,
				Msg:  "生成 token 失败: " + err.Error(),
			},
		}, nil
	}

	return &types.ForgetPasswordResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: *tokenData,
	}, nil
}
