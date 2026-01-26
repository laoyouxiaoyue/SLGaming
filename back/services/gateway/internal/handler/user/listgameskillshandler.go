// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"net/http"

	"SLGaming/back/services/gateway/internal/logic/user"
	"SLGaming/back/services/gateway/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListGameSkillsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewListGameSkillsLogic(r.Context(), svcCtx)
		resp, err := l.ListGameSkills()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
