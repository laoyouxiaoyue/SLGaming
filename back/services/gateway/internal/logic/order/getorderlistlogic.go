// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"context"
	"fmt"

	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/order/orderclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderListLogic {
	return &GetOrderListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderListLogic) GetOrderList(req *types.GetOrderListRequest) (resp *types.GetOrderListResponse, err error) {
	if l.svcCtx.OrderRPC == nil {
		return nil, fmt.Errorf("order rpc client not initialized")
	}

	// 当前登录用户（由网关鉴权中间件注入）
	userID, _ := middleware.GetUserID(l.ctx)

	var bossId, companionId uint64
	if req.Role == "companion" {
		companionId = userID
	} else {
		bossId = userID
	}

	rpcResp, err := l.svcCtx.OrderRPC.GetOrderList(l.ctx, &orderclient.GetOrderListRequest{
		BossId:      bossId,
		CompanionId: companionId,
		Status:      req.Status,
		Page:        req.Page,
		PageSize:    req.PageSize,
	})
	if err != nil {
		l.Errorf("call OrderRPC.GetOrderList failed: %v", err)
		return &types.GetOrderListResponse{
			BaseResp: types.BaseResp{
				Code: 500,
				Msg:  "获取订单列表失败: " + err.Error(),
			},
		}, nil
	}

	list := make([]types.OrderInfo, 0, len(rpcResp.Orders))
	for _, o := range rpcResp.Orders {
		list = append(list, toOrderInfo(o))
	}

	return &types.GetOrderListResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: types.GetOrderListData{
			Orders:   list,
			Total:    rpcResp.Total,
			Page:     rpcResp.Page,
			PageSize: rpcResp.PageSize,
		},
	}, nil
}
