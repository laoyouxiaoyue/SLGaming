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

	// 当前登录用户作为 operator_id
	operatorID, getUserErr := middleware.GetUserID(l.ctx)
	if getUserErr != nil || operatorID == 0 {
		l.Errorf("get user id from context failed: %v", getUserErr)
		return &types.CompleteOrderResponse{
			BaseResp: types.BaseResp{
				Code: 401,
				Msg:  "未授权",
			},
		}, nil
	}

	rpcResp, err := l.svcCtx.OrderRPC.CompleteOrder(l.ctx, &orderclient.CompleteOrderRequest{
		OrderId:    req.OrderId,
		OperatorId: operatorID,
	})
	if err != nil {
		l.Errorf("call OrderRPC.CompleteOrder failed: %v", err)
		return &types.CompleteOrderResponse{
			BaseResp: types.BaseResp{
				Code: 500,
				Msg:  "完成订单失败: " + err.Error(),
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
