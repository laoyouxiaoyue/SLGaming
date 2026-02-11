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

// AlipayNotifyHandler 支付宝支付回调
// @Summary 支付宝支付回调
// @Description 支付宝异步通知回调接口，用于处理支付结果
// @Tags 用户
// @Accept application/x-www-form-urlencoded
// @Produce text/plain
// @Param payload body map[string]string true "支付宝回调参数"
// @Success 200 {string} string "success"
// @Router /api/user/recharge/alipay/notify [post]
func AlipayNotifyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		payload := map[string]string{}
		for k, v := range r.PostForm {
			if len(v) > 0 {
				payload[k] = v[0]
			}
		}
		req := types.AlipayNotifyRequest{Payload: payload}

		l := user.NewAlipayNotifyLogic(r.Context(), svcCtx)
		resp, err := l.AlipayNotify(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			if resp != nil && resp.Code == 0 {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("success"))
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("fail"))
			}
		}
	}
}
