// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"net/http"

	"SLGaming/back/services/gateway/internal/logic/user"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetCompanionOrdersRankingHandler 获取陪玩接单数排行榜
// @Summary 获取陪玩接单数排行榜
// @Description 获取陪玩按接单数排名的排行榜
// @Tags 用户
// @Accept json
// @Produce json
// @Param page query int false "页码（从1开始）" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} types.GetCompanionOrdersRankingResponse "成功"
// @Router /api/user/companions/ranking/orders [get]
func GetCompanionOrdersRankingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetCompanionOrdersRankingRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewGetCompanionOrdersRankingLogic(r.Context(), svcCtx)
		resp, err := l.GetCompanionOrdersRanking(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			utils.WriteResponse(r.Context(), w, resp)
		}
	}
}
