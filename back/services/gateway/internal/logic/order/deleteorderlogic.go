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

type DeleteOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteOrderLogic {
	return &DeleteOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteOrderLogic) DeleteOrder(req *types.DeleteOrderRequest) (resp *types.DeleteOrderResponse, err error) {
	if l.svcCtx.OrderRPC == nil {
		return &types.DeleteOrderResponse{
			BaseResp: types.BaseResp{Code: 500, Msg: "order service unavailable"},
		}, nil
	}

	userID, err := middleware.GetUserID(l.ctx)
	if err != nil || userID == 0 {
		return &types.DeleteOrderResponse{
			BaseResp: types.BaseResp{Code: 401, Msg: "未登录或认证失败"},
		}, nil
	}

	rpcResp, err := l.svcCtx.OrderRPC.DeleteOrder(l.ctx, &orderclient.DeleteOrderRequest{
		OrderId:    req.OrderId,
		OperatorId: userID,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "DeleteOrder")
		return &types.DeleteOrderResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	return &types.DeleteOrderResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		Data: types.DeleteOrderData{
			Success: rpcResp.Success,
		},
	}, nil
}
