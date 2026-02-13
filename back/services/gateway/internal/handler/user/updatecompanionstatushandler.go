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

// UpdateCompanionStatusHandler 更新陪玩状态
// @Summary 更新陪玩状态
// @Description 更新陪玩的在线状态：0=离线, 1=在线, 2=忙碌
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body types.UpdateCompanionStatusRequest true "更新陪玩状态请求"
// @Success 200 {object} types.UpdateCompanionStatusResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/user/companion/status [put]
// @Security BearerAuth
func UpdateCompanionStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateCompanionStatusRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewUpdateCompanionStatusLogic(r.Context(), svcCtx)
		resp, err := l.UpdateCompanionStatus(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
