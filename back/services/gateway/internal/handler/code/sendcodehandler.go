// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package code

import (
	"net/http"

	"SLGaming/back/services/gateway/internal/logic/code"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/gateway/internal/validator"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// SendCodeHandler 发送验证码
// @Summary 发送验证码
// @Description 向指定手机号发送短信验证码，用于注册、登录、修改密码等场景
// @Tags 验证码
// @Accept json
// @Produce json
// @Param request body types.SendCodeRequest true "发送验证码请求"
// @Success 200 {object} types.SendCodeResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Router /api/code/send [post]
func SendCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendCodeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 表单验证
		if err := validator.ValidateSendCodeRequest(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := code.NewSendCodeLogic(r.Context(), svcCtx)
		resp, err := l.SendCode(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			// 根据响应码返回正确的 HTTP 状态码
			utils.WriteResponse(r.Context(), w, resp)
		}
	}
}
