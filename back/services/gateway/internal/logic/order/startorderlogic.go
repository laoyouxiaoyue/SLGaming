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

type StartOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStartOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartOrderLogic {
	return &StartOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StartOrderLogic) StartOrder(req *types.StartOrderRequest) (resp *types.StartOrderResponse, err error) {
	if l.svcCtx.OrderRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "OrderRPC")
		return &types.StartOrderResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	// 当前登录用户作为陪玩ID（由网关鉴权中间件注入）
	companionID, _ := middleware.GetUserID(l.ctx)

	rpcResp, err := l.svcCtx.OrderRPC.StartOrder(l.ctx, &orderclient.StartOrderRequest{
		OrderId:     req.OrderId,
		CompanionId: companionID,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "StartOrder")
		return &types.StartOrderResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	return &types.StartOrderResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: toOrderInfo(rpcResp.Order),
	}, nil
}
