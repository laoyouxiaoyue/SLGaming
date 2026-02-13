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

// ApplyCompanionHandler 申请成为陪玩
// @Summary 申请成为陪玩
// @Description 申请成为陪玩，需要设置游戏技能和价格
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body types.ApplyCompanionRequest true "申请陪玩请求"
// @Success 200 {object} types.ApplyCompanionResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/user/companion/apply [post]
// @Security BearerAuth
func ApplyCompanionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ApplyCompanionRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewApplyCompanionLogic(r.Context(), svcCtx)
		resp, err := l.ApplyCompanion(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
