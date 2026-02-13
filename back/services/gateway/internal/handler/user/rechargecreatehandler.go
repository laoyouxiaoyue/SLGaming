// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"net/http"

	"SLGaming/back/services/gateway/internal/logic/user"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// RechargeCreateHandler 创建充值订单
// @Summary 创建充值订单
// @Description 创建帅币充值订单，支持支付宝支付
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body types.RechargeCreateRequest true "创建充值订单请求"
// @Success 200 {object} types.RechargeCreateResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/user/recharge [post]
// @Security BearerAuth
func RechargeCreateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RechargeCreateRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewRechargeCreateLogic(r.Context(), svcCtx)
		resp, err := l.RechargeCreate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
