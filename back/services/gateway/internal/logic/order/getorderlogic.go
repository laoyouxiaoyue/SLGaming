package order

import (
	"context"

	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/order/orderclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderLogic) GetOrder(req *types.GetOrderRequest) (resp *types.GetOrderResponse, err error) {
	if l.svcCtx.OrderRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "OrderRPC")
		return &types.GetOrderResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	rpcResp, err := l.svcCtx.OrderRPC.GetOrder(l.ctx, &orderclient.GetOrderRequest{
		Id:      req.Id,
		OrderNo: req.OrderNo,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "GetOrder")
		return &types.GetOrderResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	if rpcResp.Order == nil {
		return &types.GetOrderResponse{
			BaseResp: types.BaseResp{
				Code: 404,
				Msg:  "订单不存在",
			},
		}, nil
	}

	return &types.GetOrderResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: toOrderInfo(rpcResp.Order),
	}, nil
}
