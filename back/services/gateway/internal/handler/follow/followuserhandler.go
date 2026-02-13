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

// FollowUserHandler 关注用户
// @Summary 关注用户
// @Description 关注指定用户，需要登录
// @Tags 关注
// @Accept json
// @Produce json
// @Param request body types.FollowUserRequest true "关注请求"
// @Success 200 {object} types.FollowUserResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/user/follow [post]
// @Security BearerAuth
func FollowUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FollowUserRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := follow.NewFollowUserLogic(r.Context(), svcCtx)
		resp, err := l.FollowUser(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
