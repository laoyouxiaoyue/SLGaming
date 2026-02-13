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

// UpdateUserHandler 更新用户信息
// @Summary 更新用户信息
// @Description 更新当前登录用户的个人信息
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body types.UpdateUserRequest true "更新用户信息请求"
// @Success 200 {object} types.UpdateUserResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/user [put]
// @Security BearerAuth
func UpdateUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateUserRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 表单验证
		if err := validator.ValidateUpdateUserRequest(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewUpdateUserLogic(r.Context(), svcCtx)
		resp, err := l.UpdateUser(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			// 根据响应码返回正确的 HTTP 状态码
			utils.WriteResponse(r.Context(), w, resp)
		}
	}
}
