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

// CreateOrderHandler 创建订单
// @Summary 创建订单
// @Description 老板向陪玩下单，创建陪玩订单
// @Tags 订单
// @Accept json
// @Produce json
// @Param request body types.CreateOrderRequest true "创建订单请求"
// @Success 200 {object} types.CreateOrderResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/order [post]
// @Security BearerAuth
func CreateOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateOrderRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := order.NewCreateOrderLogic(r.Context(), svcCtx)
		resp, err := l.CreateOrder(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			utils.WriteResponse(r.Context(), w, resp)
		}
	}
}
