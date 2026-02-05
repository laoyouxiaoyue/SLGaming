package order

import (
	"context"
	"math"

	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
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
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "OrderRPC")
		return &types.CreateOrderResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	// 从 context 中获取当前登录用户 ID（老板，由网关鉴权中间件注入）
	bossID, _ := middleware.GetUserID(l.ctx)

	if req.DurationHours <= 0 {
		return &types.CreateOrderResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "durationHours must be positive"},
		}, nil
	}

	durationMinutes64 := int64(req.DurationHours) * 60
	if durationMinutes64 > math.MaxInt32 {
		return &types.CreateOrderResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "durationHours too large"},
		}, nil
	}

	rpcReq := &orderclient.CreateOrderRequest{
		BossId:        bossID,
		CompanionId:   req.CompanionId,
		GameName:      req.GameName,
		DurationHours: req.DurationHours,
	}

	rpcResp, err := l.svcCtx.OrderRPC.CreateOrder(l.ctx, rpcReq)
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "CreateOrder")
		return &types.CreateOrderResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
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
