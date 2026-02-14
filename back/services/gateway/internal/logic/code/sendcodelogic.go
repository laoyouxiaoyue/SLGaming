// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package code

import (
	"context"

	"SLGaming/back/services/code/codeclient"
	"SLGaming/back/services/gateway/internal/helper"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendCodeLogic {
	return &SendCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendCodeLogic) SendCode(req *types.SendCodeRequest) (resp *types.SendCodeResponse, err error) {
	helper.LogRequest(l.Logger, helper.OpSendCode, map[string]interface{}{
		"phone":   helper.MaskPhone(req.Phone),
		"purpose": req.Purpose,
	})

	if l.svcCtx.CodeRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "CodeRPC")
		helper.LogError(l.Logger, helper.OpSendCode, "code rpc not available", nil, nil)
		return &types.SendCodeResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	// 调用验证码服务的 RPC
	_, err = l.svcCtx.CodeRPC.SendCode(l.ctx, &codeclient.SendCodeRequest{
		Phone:   req.Phone,
		Purpose: req.Purpose,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "SendCode")
		return &types.SendCodeResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	helper.LogSuccess(l.Logger, helper.OpSendCode, map[string]interface{}{
		"phone":   helper.MaskPhone(req.Phone),
		"purpose": req.Purpose,
	})

	return &types.SendCodeResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: types.SendCodeData{
			Success: true,
		},
	}, nil
}
