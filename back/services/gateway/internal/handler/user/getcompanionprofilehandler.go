// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"net/http"

	"SLGaming/back/services/gateway/internal/logic/user"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetCompanionProfileHandler 获取陪玩资料
// @Summary 获取陪玩资料
// @Description 获取当前登录用户的陪玩资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} types.GetCompanionProfileResponse "成功"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/user/companion/profile [get]
// @Security BearerAuth
func GetCompanionProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewGetCompanionProfileLogic(r.Context(), svcCtx)
		resp, err := l.GetCompanionProfile()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			// 根据业务 code 返回正确的 HTTP 状态码
			utils.WriteResponse(r.Context(), w, resp)
		}
	}
}
