// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCompanionProfileByIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCompanionProfileByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanionProfileByIdLogic {
	return &GetCompanionProfileByIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCompanionProfileByIdLogic) GetCompanionProfileById(req *types.GetCompanionProfileByIdRequest) (*types.GetCompanionProfileResponse, error) {
	if req.UserId == 0 {
		return &types.GetCompanionProfileResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "userId is required"},
		}, nil
	}

	if l.svcCtx.UserRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.GetCompanionProfileResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	rpcResp, err := l.svcCtx.UserRPC.GetCompanionProfile(l.ctx, &userclient.GetCompanionProfileRequest{
		UserId: req.UserId,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "GetCompanionProfileById")
		return &types.GetCompanionProfileResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	profile := rpcResp.GetProfile()
	if profile == nil {
		return &types.GetCompanionProfileResponse{
			BaseResp: types.BaseResp{Code: 404, Msg: "陪玩信息不存在"},
		}, nil
	}

	return &types.GetCompanionProfileResponse{
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
