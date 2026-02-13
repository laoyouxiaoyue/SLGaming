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

// CheckFollowStatusHandler 检查关注状态
// @Summary 检查关注状态
// @Description 检查当前用户与目标用户的关注关系状态
// @Tags 关注
// @Accept json
// @Produce json
// @Param targetUserId query uint64 true "目标用户ID"
// @Success 200 {object} types.CheckFollowStatusResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/user/follow/status [get]
// @Security BearerAuth
func CheckFollowStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CheckFollowStatusRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := follow.NewCheckFollowStatusLogic(r.Context(), svcCtx)
		resp, err := l.CheckFollowStatus(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
