// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"time"

	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RechargeListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRechargeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RechargeListLogic {
	return &RechargeListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RechargeListLogic) RechargeList(req *types.RechargeListRequest) (resp *types.RechargeListResponse, err error) {
	userID, err := middleware.GetUserID(l.ctx)
	if err != nil {
		return &types.RechargeListResponse{
			BaseResp: types.BaseResp{Code: 401, Msg: "未登录或登录已过期"},
		}, nil
	}

	if l.svcCtx.UserRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.RechargeListResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	rpcResp, err := l.svcCtx.UserRPC.RechargeList(l.ctx, &userclient.RechargeListRequest{
		UserId:   userID,
		Status:   int32(req.Status),
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "RechargeList")
		return &types.RechargeListResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	orders := make([]types.RechargeOrderInfo, 0, len(rpcResp.Orders))
	for _, o := range rpcResp.Orders {
		paidAt := ""
		if o.PaidAt > 0 {
			paidAt = time.Unix(o.PaidAt, 0).Format("2006-01-02 15:04:05")
		}
		createdAt := ""
		if o.CreatedAt > 0 {
			createdAt = time.Unix(o.CreatedAt, 0).Format("2006-01-02 15:04:05")
		}
		orders = append(orders, types.RechargeOrderInfo{
			OrderNo:   o.OrderNo,
			Status:    int(o.Status),
			Amount:    o.Amount,
			PayType:   o.PayType,
			TradeNo:   o.TradeNo,
			PaidAt:    paidAt,
			CreatedAt: createdAt,
		})
	}

	return &types.RechargeListResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		Data: types.RechargeListData{
			Orders:   orders,
			Total:    int(rpcResp.Total),
			Page:     int(rpcResp.Page),
			PageSize: int(rpcResp.PageSize),
		},
	}, nil
}
