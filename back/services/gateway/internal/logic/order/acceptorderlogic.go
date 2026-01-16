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

type AcceptOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAcceptOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AcceptOrderLogic {
	return &AcceptOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AcceptOrderLogic) AcceptOrder(req *types.AcceptOrderRequest) (resp *types.AcceptOrderResponse, err error) {
	if l.svcCtx.OrderRPC == nil {
		return nil, fmt.Errorf("order rpc client not initialized")
	}

	// 当前登录用户作为陪玩ID（由网关鉴权中间件注入）
	companionID, _ := middleware.GetUserID(l.ctx)

	rpcResp, err := l.svcCtx.OrderRPC.AcceptOrder(l.ctx, &orderclient.AcceptOrderRequest{
		OrderId:     req.OrderId,
		CompanionId: companionID,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "AcceptOrder")
		return &types.AcceptOrderResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	return &types.AcceptOrderResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: toOrderInfo(rpcResp.Order),
	}, nil
}
