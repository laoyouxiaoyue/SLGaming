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

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderRequest) (resp *types.CreateOrderResponse, err error) {
	if l.svcCtx.OrderRPC == nil {
		return nil, fmt.Errorf("order rpc client not initialized")
	}

	// 从 context 中获取当前登录用户 ID（老板）
	bossID, getUserErr := middleware.GetUserID(l.ctx)
	if getUserErr != nil || bossID == 0 {
		l.Errorf("get user id from context failed: %v", getUserErr)
		return &types.CreateOrderResponse{
			BaseResp: types.BaseResp{
				Code: 401,
				Msg:  "未授权",
			},
		}, nil
	}

	rpcReq := &orderclient.CreateOrderRequest{
		BossId:          bossID,
		CompanionId:     req.CompanionId,
		GameName:        req.GameName,
		GameMode:        req.GameMode,
		DurationMinutes: req.DurationMinutes,
	}

	rpcResp, err := l.svcCtx.OrderRPC.CreateOrder(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("call OrderRPC.CreateOrder failed: %v", err)
		return &types.CreateOrderResponse{
			BaseResp: types.BaseResp{
				Code: 500,
				Msg:  "创建订单失败: " + err.Error(),
			},
		}, nil
	}

	o := rpcResp.Order
	if o == nil {
		return &types.CreateOrderResponse{
			BaseResp: types.BaseResp{
				Code: 0,
				Msg:  "success",
			},
		}, nil
	}

	return &types.CreateOrderResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: toOrderInfo(o),
	}, nil
}
