// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"net/http"

	"SLGaming/back/services/gateway/internal/logic/order"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// RateOrderHandler 评价订单
// @Summary 评价订单
// @Description 对已完成的订单进行评价，包括评分和评论
// @Tags 订单
// @Accept json
// @Produce json
// @Param request body types.RateOrderRequest true "评价订单请求"
// @Success 200 {object} types.RateOrderResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/order/rate [post]
// @Security BearerAuth
func RateOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RateOrderRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := order.NewRateOrderLogic(r.Context(), svcCtx)
		resp, err := l.RateOrder(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			utils.WriteResponse(r.Context(), w, resp)
		}
	}
}
