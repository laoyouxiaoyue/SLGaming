// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"fmt"

	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCompanionProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCompanionProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanionProfileLogic {
	return &GetCompanionProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCompanionProfileLogic) GetCompanionProfile() (resp *types.GetCompanionProfileResponse, err error) {
	// 从 context 中获取当前登录用户 ID（由网关鉴权中间件注入）
	userID, err := middleware.GetUserID(l.ctx)
	if err != nil {
		l.Errorf("GetUserID from context failed: %v", err)
		return &types.GetCompanionProfileResponse{
			BaseResp: types.BaseResp{
				Code: 401,
				Msg:  "未登录或认证失败: " + err.Error(),
			},
		}, nil
	}
	if userID == 0 {
		l.Errorf("userID is 0, authentication may have failed")
		return &types.GetCompanionProfileResponse{
			BaseResp: types.BaseResp{
				Code: 401,
				Msg:  "未登录或认证失败",
			},
		}, nil
	}

	if l.svcCtx.UserRPC == nil {
		return nil, fmt.Errorf("user rpc client not initialized")
	}

	// 调用 User RPC 的 GetCompanionProfile 接口
	rpcResp, err := l.svcCtx.UserRPC.GetCompanionProfile(l.ctx, &userclient.GetCompanionProfileRequest{
		UserId: userID,
	})
	if err != nil {
		l.Errorf("UserRPC.GetCompanionProfile failed: %v", err)
		return &types.GetCompanionProfileResponse{
			BaseResp: types.BaseResp{
				Code: 500,
				Msg:  "获取陪玩信息失败: " + err.Error(),
			},
		}, nil
	}

	profile := rpcResp.GetProfile()
	if profile == nil {
		return &types.GetCompanionProfileResponse{
			BaseResp: types.BaseResp{
				Code: 404,
				Msg:  "陪玩信息不存在",
			},
		}, nil
	}

	return &types.GetCompanionProfileResponse{
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
