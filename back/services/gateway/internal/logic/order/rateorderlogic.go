// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"context"

	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/order/orderclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RateOrderLogic {
	return &RateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RateOrderLogic) RateOrder(req *types.RateOrderRequest) (resp *types.RateOrderResponse, err error) {
	if l.svcCtx.OrderRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "OrderRPC")
		return &types.RateOrderResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	// 当前登录用户作为 boss_id（由网关鉴权中间件注入）
	bossID, _ := middleware.GetUserID(l.ctx)

	rpcResp, err := l.svcCtx.OrderRPC.RateOrder(l.ctx, &orderclient.RateOrderRequest{
		OrderId: req.OrderId,
		BossId:  bossID,
		Rating:  req.Rating,
		Comment: req.Comment,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "RateOrder")
		return &types.RateOrderResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	return &types.RateOrderResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: toOrderInfo(rpcResp.Order),
	}, nil
}
