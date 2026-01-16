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

type CancelOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelOrderLogic) CancelOrder(req *types.CancelOrderRequest) (resp *types.CancelOrderResponse, err error) {
	if l.svcCtx.OrderRPC == nil {
		return nil, fmt.Errorf("order rpc client not initialized")
	}

	// 当前登录用户作为 operator_id（由网关鉴权中间件注入）
	operatorID, _ := middleware.GetUserID(l.ctx)

	rpcResp, err := l.svcCtx.OrderRPC.CancelOrder(l.ctx, &orderclient.CancelOrderRequest{
		OrderId:    req.OrderId,
		OperatorId: operatorID,
		Reason:     req.Reason,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "CancelOrder")
		return &types.CancelOrderResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	return &types.CancelOrderResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: toOrderInfo(rpcResp.Order),
	}, nil
}
