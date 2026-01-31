// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"strings"

	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadAvatarLogic {
	return &UploadAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadAvatarLogic) UploadAvatar(req *types.UploadAvatarRequest) (resp *types.UploadAvatarResponse, err error) {
	// 从 context 中获取当前登录用户 ID
	currentUserID, err := middleware.GetUserID(l.ctx)
	if err != nil {
		return &types.UploadAvatarResponse{
			BaseResp: types.BaseResp{
				Code: 401,
				Msg:  "未登录或登录已过期",
			},
		}, nil
	}

	avatarUrl := strings.TrimSpace(req.Avatar)
	if avatarUrl == "" {
		return &types.UploadAvatarResponse{
			BaseResp: types.BaseResp{
				Code: 400,
				Msg:  "头像地址不能为空",
			},
		}, nil
	}

	if l.svcCtx.UserRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.UploadAvatarResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	_, err = l.svcCtx.UserRPC.UpdateUser(l.ctx, &userclient.UpdateUserRequest{
		Id:        currentUserID,
		AvatarUrl: avatarUrl,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "UpdateUser")
		return &types.UploadAvatarResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	return &types.UploadAvatarResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: types.UploadAvatarData{
			AvatarUrl: avatarUrl,
		},
	}, nil
}
