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

// UpdateCompanionProfileHandler 更新陪玩资料
// @Summary 更新陪玩资料
// @Description 更新当前登录用户的陪玩资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body types.UpdateCompanionProfileRequest true "更新陪玩资料请求"
// @Success 200 {object} types.UpdateCompanionProfileResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/user/companion/profile [put]
// @Security BearerAuth
func UpdateCompanionProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateCompanionProfileRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewUpdateCompanionProfileLogic(r.Context(), svcCtx)
		resp, err := l.UpdateCompanionProfile(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			// 根据业务 code 返回正确的 HTTP 状态码
			utils.WriteResponse(r.Context(), w, resp)
		}
	}
}
