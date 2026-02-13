// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package agent

import (
	"net/http"

	"SLGaming/back/services/gateway/internal/logic/agent"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// RecommendCompanionHandler AI推荐陪玩
// @Summary AI推荐陪玩
// @Description 根据用户需求，使用AI智能推荐合适的陪玩
// @Tags AI助手
// @Accept json
// @Produce json
// @Param request body types.RecommendCompanionRequest true "AI推荐陪玩请求"
// @Success 200 {object} types.RecommendCompanionResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/agent/recommend [post]
// @Security BearerAuth
func RecommendCompanionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RecommendCompanionRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := agent.NewRecommendCompanionLogic(r.Context(), svcCtx)
		resp, err := l.RecommendCompanion(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			utils.WriteResponse(r.Context(), w, resp)
		}
	}
}
