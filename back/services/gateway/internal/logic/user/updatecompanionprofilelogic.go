// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"fmt"

	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCompanionProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCompanionProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCompanionProfileLogic {
	return &UpdateCompanionProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateCompanionProfileLogic) UpdateCompanionProfile(req *types.UpdateCompanionProfileRequest) (resp *types.UpdateCompanionProfileResponse, err error) {
	// 从 context 中获取当前登录用户 ID（由网关鉴权中间件注入）
	l.Infof("UpdateCompanionProfile: attempting to get userID from context")
	userID, err := middleware.GetUserID(l.ctx)
	if err != nil {
		code, msg := utils.HandleError(err, l.Logger, "GetUserID")
		return &types.UpdateCompanionProfileResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}
	l.Infof("UpdateCompanionProfile: extracted userID=%d from context", userID)
	if userID == 0 {
		l.Errorf("userID is 0, authentication may have failed")
		return &types.UpdateCompanionProfileResponse{
			BaseResp: types.BaseResp{
				Code: 401,
				Msg:  "未登录或认证失败",
			},
		}, nil
	}

	if l.svcCtx.UserRPC == nil {
		return nil, fmt.Errorf("user rpc client not initialized")
	}

	// 调用 User RPC 的 UpdateCompanionProfile 接口
	l.Infof("UpdateCompanionProfile: calling RPC with userID=%d, pricePerHour=%d, status=%d", userID, req.PricePerHour, req.Status)
	rpcResp, err := l.svcCtx.UserRPC.UpdateCompanionProfile(l.ctx, &userclient.UpdateCompanionProfileRequest{
		UserId:       userID,
		GameSkills:   req.GameSkills,
		PricePerHour: req.PricePerHour,
		Status:       int32(req.Status),
	})
	l.Infof("UpdateCompanionProfile: RPC call completed, err=%v", err)
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "UpdateCompanionProfile")
		return &types.UpdateCompanionProfileResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	profile := rpcResp.GetProfile()
	if profile == nil {
		return &types.UpdateCompanionProfileResponse{
			BaseResp: types.BaseResp{
				Code: 404,
				Msg:  "陪玩信息不存在",
			},
		}, nil
	}

	return &types.UpdateCompanionProfileResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: types.CompanionInfo{
			UserId:       profile.UserId,
			GameSkills:   profile.GameSkills,
			PricePerHour: profile.PricePerHour,
			Status:       int(profile.Status),
			Rating:       profile.Rating,
			TotalOrders:  profile.TotalOrders,
			IsVerified:   profile.IsVerified,
		},
	}, nil
}
