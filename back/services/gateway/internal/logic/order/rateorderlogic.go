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
		return nil, fmt.Errorf("order rpc client not initialized")
	}

	// 当前登录用户作为 boss_id
	bossID, getUserErr := middleware.GetUserID(l.ctx)
	if getUserErr != nil || bossID == 0 {
		l.Errorf("get user id from context failed: %v", getUserErr)
		return &types.RateOrderResponse{
			BaseResp: types.BaseResp{
				Code: 401,
				Msg:  "未授权",
			},
		}, nil
	}

	rpcResp, err := l.svcCtx.OrderRPC.RateOrder(l.ctx, &orderclient.RateOrderRequest{
		OrderId: req.OrderId,
		BossId:  bossID,
		Rating:  req.Rating,
		Comment: req.Comment,
	})
	if err != nil {
		l.Errorf("call OrderRPC.RateOrder failed: %v", err)
		return &types.RateOrderResponse{
			BaseResp: types.BaseResp{
				Code: 500,
				Msg:  "评价订单失败: " + err.Error(),
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
