// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"net/http"

	"SLGaming/back/services/gateway/internal/logic/order"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// DeleteOrderHandler 删除订单
// @Summary 删除订单
// @Description 删除已完成、已取消或已评价的订单（软删除，仅对当前用户隐藏）
// @Tags 订单
// @Accept json
// @Produce json
// @Param request body types.DeleteOrderRequest true "删除订单请求"
// @Success 200 {object} types.DeleteOrderResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Failure 403 {object} types.BaseResp "权限不足"
// @Failure 404 {object} types.BaseResp "订单不存在"
// @Router /api/order/delete [post]
// @Security BearerAuth
func DeleteOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteOrderRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := order.NewDeleteOrderLogic(r.Context(), svcCtx)
		resp, err := l.DeleteOrder(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
