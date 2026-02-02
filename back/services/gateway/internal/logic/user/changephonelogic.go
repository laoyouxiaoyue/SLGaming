// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"strings"

	"SLGaming/back/services/code/codeclient"
	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangePhoneLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangePhoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePhoneLogic {
	return &ChangePhoneLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangePhoneLogic) ChangePhone(req *types.ChangePhoneRequest) (resp *types.ChangePhoneResponse, err error) {
	userID, err := middleware.GetUserID(l.ctx)
	if err != nil {
		return &types.ChangePhoneResponse{
			BaseResp: types.BaseResp{Code: 401, Msg: "未登录或登录已过期"},
		}, nil
	}

	oldPhone := strings.TrimSpace(req.OldPhone)
	newPhone := strings.TrimSpace(req.NewPhone)
	oldCode := strings.TrimSpace(req.OldCode)
	newCode := strings.TrimSpace(req.NewCode)
	if oldPhone == "" || newPhone == "" || newCode == "" {
		return &types.ChangePhoneResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "手机号或验证码不能为空"},
		}, nil
	}
	if oldPhone == newPhone {
		return &types.ChangePhoneResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "新旧手机号不能相同"},
		}, nil
	}

	if l.svcCtx.UserRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.ChangePhoneResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	// 校验原手机号是否为当前用户
	userResp, err := l.svcCtx.UserRPC.GetUser(l.ctx, &userclient.GetUserRequest{Id: userID})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "GetUser")
		return &types.ChangePhoneResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}
	if userResp.User == nil || strings.TrimSpace(userResp.User.Phone) != oldPhone {
		return &types.ChangePhoneResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "原手机号不匹配"},
		}, nil
	}

	// 验证原手机号验证码
	if l.svcCtx.CodeRPC != nil {
		verifyResp, err := l.svcCtx.CodeRPC.VerifyCode(l.ctx, &codeclient.VerifyCodeRequest{
			Phone:   oldPhone,
			Purpose: "change_phone",
			Code:    oldCode,
		})
		if err != nil {
			code, msg := utils.HandleRPCError(err, l.Logger, "VerifyCode")
			return &types.ChangePhoneResponse{
				BaseResp: types.BaseResp{Code: code, Msg: msg},
			}, nil
		}
		if !verifyResp.Passed {
			return &types.ChangePhoneResponse{
				BaseResp: types.BaseResp{Code: 400, Msg: "验证码错误或已过期"},
			}, nil
		}
	}

	// 验证新手机号验证码
	if l.svcCtx.CodeRPC != nil {
		verifyResp, err := l.svcCtx.CodeRPC.VerifyCode(l.ctx, &codeclient.VerifyCodeRequest{
			Phone:   newPhone,
			Purpose: "change_phone_new",
			Code:    newCode,
		})
		if err != nil {
			code, msg := utils.HandleRPCError(err, l.Logger, "VerifyCode")
			return &types.ChangePhoneResponse{
				BaseResp: types.BaseResp{Code: code, Msg: msg},
			}, nil
		}
		if !verifyResp.Passed {
			return &types.ChangePhoneResponse{
				BaseResp: types.BaseResp{Code: 400, Msg: "验证码错误或已过期"},
			}, nil
		}
	}

	_, err = l.svcCtx.UserRPC.ChangePhone(l.ctx, &userclient.ChangePhoneRequest{
		UserId:   userID,
		OldPhone: oldPhone,
		NewPhone: newPhone,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "ChangePhone")
		return &types.ChangePhoneResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	return &types.ChangePhoneResponse{BaseResp: types.BaseResp{Code: 0, Msg: "success"}}, nil
}
