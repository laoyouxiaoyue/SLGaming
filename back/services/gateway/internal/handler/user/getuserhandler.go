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

// GetUserHandler 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户或指定用户的详细信息
// @Tags 用户
// @Accept json
// @Produce json
// @Param id query uint64 false "用户ID（可选，不传则获取当前用户）"
// @Param uid query uint64 false "用户UID（可选）"
// @Param phone query string false "手机号（可选）"
// @Success 200 {object} types.GetUserResponse "成功"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/user [get]
// @Security BearerAuth
func GetUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetUserRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewGetUserLogic(r.Context(), svcCtx)
		resp, err := l.GetUser(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			// 根据响应码返回正确的 HTTP 状态码
			utils.WriteResponse(r.Context(), w, resp)
		}
	}
}
