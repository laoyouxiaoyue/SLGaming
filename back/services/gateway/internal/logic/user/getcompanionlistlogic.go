// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"fmt"

	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCompanionListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCompanionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanionListLogic {
	return &GetCompanionListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCompanionListLogic) GetCompanionList(req *types.GetCompanionListRequest) (resp *types.GetCompanionListResponse, err error) {
	if l.svcCtx.UserRPC == nil {
		return nil, fmt.Errorf("user rpc client not initialized")
	}

	// 调用 User RPC 的 GetCompanionList 接口
	// 注意：如果 status 为 0，RPC 层会将其视为"未指定"并使用默认值（在线）
	// 如果用户明确想查询 status=0（离线），需要确保 gateway 层正确传递
	rpcResp, err := l.svcCtx.UserRPC.GetCompanionList(l.ctx, &userclient.GetCompanionListRequest{
		GameSkill:  req.GameSkill,
		MinPrice:   int32(req.MinPrice),
		MaxPrice:   int32(req.MaxPrice),
		Status:     int32(req.Status),
		IsVerified: req.IsVerified,
		Page:       int32(req.Page),
		PageSize:   int32(req.PageSize),
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "GetCompanionList")
		return &types.GetCompanionListResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 转换陪玩列表
	companions := make([]types.CompanionInfo, 0, len(rpcResp.Companions))
	for _, cp := range rpcResp.Companions {
		companions = append(companions, types.CompanionInfo{
			UserId:       cp.UserId,
			GameSkill:    cp.GameSkill,
			PricePerHour: cp.PricePerHour,
			Status:       int(cp.Status),
			Rating:       cp.Rating,
			TotalOrders:  cp.TotalOrders,
			IsVerified:   cp.IsVerified,
			AvatarUrl:    cp.AvatarUrl,
			Bio:          cp.Bio,
		})
	}

	return &types.GetCompanionListResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: types.GetCompanionListData{
			Companions: companions,
			Total:      int(rpcResp.Total),
			Page:       int(rpcResp.Page),
			PageSize:   int(rpcResp.PageSize),
		},
	}, nil
}
