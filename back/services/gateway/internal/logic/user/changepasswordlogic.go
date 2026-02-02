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

type ChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangePasswordLogic) ChangePassword(req *types.ChangePasswordRequest) (resp *types.ChangePasswordResponse, err error) {
	userID, err := middleware.GetUserID(l.ctx)
	if err != nil {
		return &types.ChangePasswordResponse{
			BaseResp: types.BaseResp{Code: 401, Msg: "未登录或登录已过期"},
		}, nil
	}

	oldPhone := strings.TrimSpace(req.OldPhone)
	oldCode := strings.TrimSpace(req.OldCode)
	newPassword := strings.TrimSpace(req.NewPassword)
	if oldPhone == "" || newPassword == "" {
		return &types.ChangePasswordResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "手机号和新密码不能为空"},
		}, nil
	}

	if l.svcCtx.UserRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.ChangePasswordResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	// 校验原手机号是否为当前用户
	userResp, err := l.svcCtx.UserRPC.GetUser(l.ctx, &userclient.GetUserRequest{Id: userID})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "GetUser")
		return &types.ChangePasswordResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}
	if userResp.User == nil || strings.TrimSpace(userResp.User.Phone) != oldPhone {
		return &types.ChangePasswordResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "原手机号不匹配"},
		}, nil
	}

	// 验证原手机号验证码
	if l.svcCtx.CodeRPC != nil {
		verifyResp, err := l.svcCtx.CodeRPC.VerifyCode(l.ctx, &codeclient.VerifyCodeRequest{
			Phone:   oldPhone,
			Purpose: "change_password",
			Code:    oldCode,
		})
		if err != nil {
			code, msg := utils.HandleRPCError(err, l.Logger, "VerifyCode")
			return &types.ChangePasswordResponse{
				BaseResp: types.BaseResp{Code: code, Msg: msg},
			}, nil
		}
		if !verifyResp.Passed {
			return &types.ChangePasswordResponse{
				BaseResp: types.BaseResp{Code: 400, Msg: "验证码错误或已过期"},
			}, nil
		}
	}

	_, err = l.svcCtx.UserRPC.ChangePassword(l.ctx, &userclient.ChangePasswordRequest{
		UserId:      userID,
		OldPhone:    oldPhone,
		NewPassword: newPassword,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "ChangePassword")
		return &types.ChangePasswordResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	return &types.ChangePasswordResponse{BaseResp: types.BaseResp{Code: 0, Msg: "success"}}, nil
}
