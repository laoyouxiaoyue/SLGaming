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

// ChangePhoneHandler 更换手机号
// @Summary 更换手机号
// @Description 更换当前登录用户的手机号，需要验证旧手机号和新手机号的验证码
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body types.ChangePhoneRequest true "更换手机号请求"
// @Success 200 {object} types.ChangePhoneResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/user/change-phone [put]
// @Security BearerAuth
func ChangePhoneHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChangePhoneRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewChangePhoneLogic(r.Context(), svcCtx)
		resp, err := l.ChangePhone(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
