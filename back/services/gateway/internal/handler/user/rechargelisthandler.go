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

// RechargeListHandler 获取充值记录列表
// @Summary 获取充值记录列表
// @Description 获取当前用户的充值记录列表，支持分页
// @Tags 用户
// @Accept json
// @Produce json
// @Param status query int false "订单状态筛选"
// @Param page query int false "页码（从1开始）" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} types.RechargeListResponse "成功"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/user/recharge/list [get]
// @Security BearerAuth
func RechargeListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RechargeListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewRechargeListLogic(r.Context(), svcCtx)
		resp, err := l.RechargeList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
