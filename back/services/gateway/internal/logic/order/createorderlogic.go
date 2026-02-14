package order

import (
	"context"

	"SLGaming/back/services/gateway/internal/helper"
	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/order/orderclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderRequest) (resp *types.CreateOrderResponse, err error) {
	// 从 context 中获取当前登录用户 ID（老板，由网关鉴权中间件注入）
	bossID, _ := middleware.GetUserID(l.ctx)

	helper.LogRequest(l.Logger, helper.OpCreateOrder, map[string]interface{}{
		"boss_id":        bossID,
		"companion_id":   req.CompanionId,
		"game_name":      req.GameName,
		"duration_hours": req.DurationHours,
	})

	if l.svcCtx.OrderRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "OrderRPC")
		helper.LogError(l.Logger, helper.OpCreateOrder, "order rpc not available", nil, nil)
		return &types.CreateOrderResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	if req.DurationHours <= 0 {
		helper.LogWarning(l.Logger, helper.OpCreateOrder, "invalid duration hours", map[string]interface{}{
			"duration_hours": req.DurationHours,
		})
		return &types.CreateOrderResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "durationHours must be positive"},
		}, nil
	}

	rpcReq := &orderclient.CreateOrderRequest{
		BossId:        bossID,
		CompanionId:   req.CompanionId,
		GameName:      req.GameName,
		DurationHours: req.DurationHours,
	}

	rpcResp, err := l.svcCtx.OrderRPC.CreateOrder(l.ctx, rpcReq)
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "CreateOrder")
		return &types.CreateOrderResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	o := rpcResp.Order
	if o == nil {
		helper.LogSuccess(l.Logger, helper.OpCreateOrder, map[string]interface{}{
			"boss_id":  bossID,
			"order_id": 0,
		})
		return &types.CreateOrderResponse{
			BaseResp: types.BaseResp{
				Code: 0,
				Msg:  "success",
			},
		}, nil
	}

	helper.LogSuccess(l.Logger, helper.OpCreateOrder, map[string]interface{}{
		"order_id":     o.Id,
		"order_no":     o.OrderNo,
		"boss_id":      o.BossId,
		"companion_id": o.CompanionId,
		"total_amount": o.TotalAmount,
	})

	return &types.CreateOrderResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: toOrderInfo(o),
	}, nil
}
