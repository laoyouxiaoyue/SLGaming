// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"context"
	"fmt"

	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/order/orderclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CompleteOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCompleteOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompleteOrderLogic {
	return &CompleteOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CompleteOrderLogic) CompleteOrder(req *types.CompleteOrderRequest) (resp *types.CompleteOrderResponse, err error) {
	if l.svcCtx.OrderRPC == nil {
		return nil, fmt.Errorf("order rpc client not initialized")
	}

	// 当前登录用户作为 operator_id（由网关鉴权中间件注入）
	operatorID, _ := middleware.GetUserID(l.ctx)

	rpcResp, err := l.svcCtx.OrderRPC.CompleteOrder(l.ctx, &orderclient.CompleteOrderRequest{
		OrderId:    req.OrderId,
		OperatorId: operatorID,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "CompleteOrder")
		return &types.CompleteOrderResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	return &types.CompleteOrderResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: toOrderInfo(rpcResp.Order),
	}, nil
}
