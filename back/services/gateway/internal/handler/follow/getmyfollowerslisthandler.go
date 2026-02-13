// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package follow

import (
	"net/http"

	"SLGaming/back/services/gateway/internal/logic/follow"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetMyFollowersListHandler 获取我的粉丝列表
// @Summary 获取我的粉丝列表
// @Description 获取当前登录用户的粉丝列表，支持分页
// @Tags 关注
// @Accept json
// @Produce json
// @Param page query int false "页码（从1开始）" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} types.GetMyFollowersListResponse "成功"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/user/followers [get]
// @Security BearerAuth
func GetMyFollowersListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetMyFollowersListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := follow.NewGetMyFollowersListLogic(r.Context(), svcCtx)
		resp, err := l.GetMyFollowersList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
