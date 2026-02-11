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

// AcceptOrderHandler 接单
// @Summary 接单
// @Description 陪玩接单，将订单状态从待接单改为服务中
// @Tags 订单
// @Accept json
// @Produce json
// @Param request body types.AcceptOrderRequest true "接单请求"
// @Success 200 {object} types.AcceptOrderResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/order/accept [put]
// @Security BearerAuth
func AcceptOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AcceptOrderRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := order.NewAcceptOrderLogic(r.Context(), svcCtx)
		resp, err := l.AcceptOrder(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			utils.WriteResponse(r.Context(), w, resp)
		}
	}
}
