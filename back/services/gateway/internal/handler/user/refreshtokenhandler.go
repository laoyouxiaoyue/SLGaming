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

// RefreshTokenHandler 刷新 Token
// @Summary 刷新 Token
// @Description 使用 Refresh Token 获取新的 Access Token
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body types.RefreshTokenRequest true "刷新 Token 请求"
// @Success 200 {object} types.RefreshTokenResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Router /api/user/refresh-token [post]
func RefreshTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RefreshTokenRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewRefreshTokenLogic(r.Context(), svcCtx)
		resp, err := l.RefreshToken(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			// 将 token 设置到响应头
			if resp != nil && resp.Data.AccessToken != "" {
				w.Header().Set("Authorization", "Bearer "+resp.Data.AccessToken)
			}
			if resp != nil && resp.Data.RefreshToken != "" {
				w.Header().Set("X-Refresh-Token", resp.Data.RefreshToken)
			}
			// 根据响应码返回正确的 HTTP 状态码
			utils.WriteResponse(r.Context(), w, resp)
		}
	}
}
