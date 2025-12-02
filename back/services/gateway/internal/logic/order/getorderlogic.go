package order

import (
	"context"
	"fmt"

	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
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
		return nil, fmt.Errorf("order rpc client not initialized")
	}

	rpcResp, err := l.svcCtx.OrderRPC.GetOrder(l.ctx, &orderclient.GetOrderRequest{
		Id:      req.Id,
		OrderNo: req.OrderNo,
	})
	if err != nil {
		l.Errorf("call OrderRPC.GetOrder failed: %v", err)
		return &types.GetOrderResponse{
			BaseResp: types.BaseResp{
				Code: 500,
				Msg:  "获取订单详情失败: " + err.Error(),
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
