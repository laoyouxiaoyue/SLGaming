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

// UnfollowUserHandler 取消关注
// @Summary 取消关注
// @Description 取消关注指定用户，需要登录
// @Tags 关注
// @Accept json
// @Produce json
// @Param request body types.UnfollowUserRequest true "取消关注请求"
// @Success 200 {object} types.UnfollowUserResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/user/unfollow [post]
// @Security BearerAuth
func UnfollowUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UnfollowUserRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := follow.NewUnfollowUserLogic(r.Context(), svcCtx)
		resp, err := l.UnfollowUser(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
