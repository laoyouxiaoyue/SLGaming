// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	companionStatusOffline = 0
	companionStatusOnline  = 1
	companionStatusBusy    = 2
)

type UpdateCompanionStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCompanionStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCompanionStatusLogic {
	return &UpdateCompanionStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateCompanionStatusLogic) UpdateCompanionStatus(req *types.UpdateCompanionStatusRequest) (resp *types.UpdateCompanionStatusResponse, err error) {
	userID, err := middleware.GetUserID(l.ctx)
	if err != nil {
		return &types.UpdateCompanionStatusResponse{
			BaseResp: types.BaseResp{Code: 401, Msg: "未登录或登录已过期"},
		}, nil
	}

	if req.Status != companionStatusOffline && req.Status != companionStatusOnline && req.Status != companionStatusBusy {
		return &types.UpdateCompanionStatusResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "状态参数无效"},
		}, nil
	}

	if l.svcCtx.UserRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.UpdateCompanionStatusResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	role, roleErr := middleware.GetUserRole(l.ctx)
	if roleErr != nil || role == 0 {
		roleResp, err := l.svcCtx.UserRPC.GetUser(l.ctx, &userclient.GetUserRequest{Id: userID})
		if err != nil {
			code, msg := utils.HandleRPCError(err, l.Logger, "GetUser")
			return &types.UpdateCompanionStatusResponse{
				BaseResp: types.BaseResp{Code: code, Msg: msg},
			}, nil
		}
		if roleResp != nil && roleResp.User != nil {
			role = roleResp.User.Role
		}
	}
	if role != roleCompanion {
		return &types.UpdateCompanionStatusResponse{
			BaseResp: types.BaseResp{Code: 403, Msg: "仅陪玩可操作状态"},
		}, nil
	}

	// 忙碌状态禁止切换为在线/离线
	currentProfile, err := l.svcCtx.UserRPC.GetCompanionProfile(l.ctx, &userclient.GetCompanionProfileRequest{
		UserId: userID,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "GetCompanionProfile")
		return &types.UpdateCompanionStatusResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}
	if currentProfile.GetProfile() != nil && currentProfile.GetProfile().Status == companionStatusBusy && req.Status != companionStatusBusy {
		return &types.UpdateCompanionStatusResponse{
			BaseResp: types.BaseResp{Code: 409, Msg: "忙碌状态下禁止上下线"},
		}, nil
	}

	rpcResp, err := l.svcCtx.UserRPC.UpdateCompanionProfile(l.ctx, &userclient.UpdateCompanionProfileRequest{
		UserId: userID,
		Status: int32(req.Status),
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "UpdateCompanionProfile")
		return &types.UpdateCompanionStatusResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	profile := rpcResp.GetProfile()
	if profile == nil {
		return &types.UpdateCompanionStatusResponse{
			BaseResp: types.BaseResp{Code: 404, Msg: "陪玩信息不存在"},
		}, nil
	}

	return &types.UpdateCompanionStatusResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		Data: types.CompanionInfo{
			UserId:       profile.UserId,
			GameSkill:    profile.GameSkill,
			PricePerHour: profile.PricePerHour,
			Status:       int(profile.Status),
			Rating:       profile.Rating,
			TotalOrders:  profile.TotalOrders,
			IsVerified:   profile.IsVerified,
			Nickname:     profile.Nickname,
			AvatarUrl:    profile.AvatarUrl,
			Bio:          profile.Bio,
		},
	}, nil
}
