// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"fmt"

	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
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
	rpcResp, err := l.svcCtx.UserRPC.GetCompanionList(l.ctx, &userclient.GetCompanionListRequest{
		GameSkills: req.GameSkills,
		MinPrice:   int32(req.MinPrice),
		MaxPrice:   int32(req.MaxPrice),
		Status:     int32(req.Status),
		IsVerified: req.IsVerified,
		Page:       int32(req.Page),
		PageSize:   int32(req.PageSize),
	})
	if err != nil {
		l.Errorf("UserRPC.GetCompanionList failed: %v", err)
		return &types.GetCompanionListResponse{
			BaseResp: types.BaseResp{
				Code: 500,
				Msg:  "获取陪玩列表失败: " + err.Error(),
			},
		}, nil
	}

	// 转换陪玩列表
	companions := make([]types.CompanionInfo, 0, len(rpcResp.Companions))
	for _, cp := range rpcResp.Companions {
		companions = append(companions, types.CompanionInfo{
			UserId:       cp.UserId,
			GameSkills:   cp.GameSkills,
			PricePerHour: cp.PricePerHour,
			Status:       int(cp.Status),
			Rating:       cp.Rating,
			TotalOrders:  cp.TotalOrders,
			IsVerified:   cp.IsVerified,
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
