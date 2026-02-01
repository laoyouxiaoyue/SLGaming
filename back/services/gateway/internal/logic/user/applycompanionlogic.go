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

const (
	roleBoss      = 1
	roleCompanion = 2
	roleAdmin     = 3
)

type ApplyCompanionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApplyCompanionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyCompanionLogic {
	return &ApplyCompanionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApplyCompanionLogic) ApplyCompanion(req *types.ApplyCompanionRequest) (resp *types.ApplyCompanionResponse, err error) {
	userID, err := middleware.GetUserID(l.ctx)
	if err != nil {
		return &types.ApplyCompanionResponse{
			BaseResp: types.BaseResp{Code: 401, Msg: "未登录或登录已过期"},
		}, nil
	}

	if l.svcCtx.UserRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.ApplyCompanionResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	gameSkill := strings.TrimSpace(req.GameSkill)
	if gameSkill == "" {
		return &types.ApplyCompanionResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "游戏技能不能为空"},
		}, nil
	}
	if req.PricePerHour <= 0 {
		return &types.ApplyCompanionResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "每小时价格必须大于0"},
		}, nil
	}

	currentRole, roleErr := middleware.GetUserRole(l.ctx)
	if roleErr != nil || currentRole == 0 {
		roleResp, err := l.svcCtx.UserRPC.GetUser(l.ctx, &userclient.GetUserRequest{Id: userID})
		if err != nil {
			code, msg := utils.HandleRPCError(err, l.Logger, "GetUser")
			return &types.ApplyCompanionResponse{
				BaseResp: types.BaseResp{Code: code, Msg: msg},
			}, nil
		}
		if roleResp != nil && roleResp.User != nil {
			currentRole = roleResp.User.Role
		}
	}

	if currentRole == roleAdmin {
		return &types.ApplyCompanionResponse{
			BaseResp: types.BaseResp{Code: 403, Msg: "管理员无需申请"},
		}, nil
	}
	if currentRole != roleBoss && currentRole != roleCompanion {
		return &types.ApplyCompanionResponse{
			BaseResp: types.BaseResp{Code: 403, Msg: "仅老板可申请成为陪玩"},
		}, nil
	}

	// 如果是老板，先升级为陪玩
	if currentRole == roleBoss {
		_, err = l.svcCtx.UserRPC.UpdateUser(l.ctx, &userclient.UpdateUserRequest{
			Id:   userID,
			Role: roleCompanion,
			Bio:  req.Bio,
		})
		if err != nil {
			code, msg := utils.HandleRPCError(err, l.Logger, "UpdateUser")
			return &types.ApplyCompanionResponse{
				BaseResp: types.BaseResp{Code: code, Msg: msg},
			}, nil
		}
	} else if strings.TrimSpace(req.Bio) != "" {
		_, _ = l.svcCtx.UserRPC.UpdateUser(l.ctx, &userclient.UpdateUserRequest{
			Id:  userID,
			Bio: req.Bio,
		})
	}

	// 设置陪玩信息（默认离线）
	_, err = l.svcCtx.UserRPC.UpdateCompanionProfile(l.ctx, &userclient.UpdateCompanionProfileRequest{
		UserId:       userID,
		GameSkill:    gameSkill,
		PricePerHour: req.PricePerHour,
		Status:       0,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "UpdateCompanionProfile")
		return &types.ApplyCompanionResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	userResp, err := l.svcCtx.UserRPC.GetUser(l.ctx, &userclient.GetUserRequest{Id: userID})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "GetUser")
		return &types.ApplyCompanionResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	return &types.ApplyCompanionResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		Data: types.UserInfo{
			Id:        userResp.User.Id,
			Uid:       userResp.User.Uid,
			Nickname:  userResp.User.Nickname,
			Phone:     userResp.User.Phone,
			Role:      int(userResp.User.Role),
			AvatarUrl: userResp.User.AvatarUrl,
			Bio:       userResp.User.Bio,
		},
	}, nil
}
