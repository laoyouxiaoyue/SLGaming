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

// GetCompanionListHandler 获取陪玩列表
// @Summary 获取陪玩列表
// @Description 获取陪玩列表，支持按游戏技能、价格、状态等筛选
// @Tags 用户
// @Accept json
// @Produce json
// @Param gameSkill query string false "游戏技能筛选"
// @Param minPrice query int false "最低价格"
// @Param maxPrice query int false "最高价格"
// @Param status query int false "状态筛选：0=离线, 1=在线, 2=忙碌" default(1)
// @Param isVerified query bool false "是否只返回认证陪玩"
// @Param page query int false "页码（从1开始）" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} types.GetCompanionListResponse "成功"
// @Router /api/user/companions [get]
func GetCompanionListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetCompanionListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewGetCompanionListLogic(r.Context(), svcCtx)
		resp, err := l.GetCompanionList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			// 根据业务 code 返回正确的 HTTP 状态码
			utils.WriteResponse(r.Context(), w, resp)
		}
	}
}
