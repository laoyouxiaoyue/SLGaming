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

// StartOrderHandler 开始服务
// @Summary 开始服务
// @Description 开始服务订单，将订单状态从待支付改为服务中
// @Tags 订单
// @Accept json
// @Produce json
// @Param request body types.StartOrderRequest true "开始服务请求"
// @Success 200 {object} types.StartOrderResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/order/start [put]
// @Security BearerAuth
func StartOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.StartOrderRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := order.NewStartOrderLogic(r.Context(), svcCtx)
		resp, err := l.StartOrder(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			utils.WriteResponse(r.Context(), w, resp)
		}
	}
}
