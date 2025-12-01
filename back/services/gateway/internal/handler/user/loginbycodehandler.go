// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"net/http"

	"SLGaming/back/services/gateway/internal/logic/user"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/gateway/internal/validator"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func LoginByCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginByCodeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 表单验证
		if err := validator.ValidateLoginByCodeRequest(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewLoginByCodeLogic(r.Context(), svcCtx)
		resp, err := l.LoginByCode(&req)
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
