// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"strings"

	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type RechargeQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRechargeQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RechargeQueryLogic {
	return &RechargeQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RechargeQueryLogic) RechargeQuery(req *types.RechargeQueryRequest) (resp *types.RechargeQueryResponse, err error) {
	userID, err := middleware.GetUserID(l.ctx)
	if err != nil {
		return &types.RechargeQueryResponse{
			BaseResp: types.BaseResp{Code: 401, Msg: "未登录或登录已过期"},
		}, nil
	}

	orderNo := strings.TrimSpace(req.OrderNo)
	if orderNo == "" {
		return &types.RechargeQueryResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "充值单号不能为空"},
		}, nil
	}

	order, err := loadRechargeOrder(l.svcCtx.CacheRedis, orderNo)
	if err != nil {
		code, msg := utils.HandleError(err, l.Logger, "LoadRechargeOrder")
		return &types.RechargeQueryResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}
	if order.UserId != userID {
		return &types.RechargeQueryResponse{
			BaseResp: types.BaseResp{Code: 403, Msg: "无权查询该订单"},
		}, nil
	}

	return &types.RechargeQueryResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		Data: types.RechargeQueryData{
			OrderNo: order.OrderNo,
			Status:  order.Status,
			Amount:  order.Amount,
		},
	}, nil
}
