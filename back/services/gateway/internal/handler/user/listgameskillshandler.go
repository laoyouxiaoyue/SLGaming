// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"net/http"

	"SLGaming/back/services/gateway/internal/logic/user"
	"SLGaming/back/services/gateway/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// ListGameSkillsHandler 获取游戏技能列表
// @Summary 获取游戏技能列表
// @Description 获取所有支持的游戏技能列表
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} types.ListGameSkillsResponse "成功"
// @Router /api/user/gameskills [get]
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
