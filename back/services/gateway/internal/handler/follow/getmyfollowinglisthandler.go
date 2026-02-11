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

// GetMyFollowingListHandler 获取我的关注列表
// @Summary 获取我的关注列表
// @Description 获取当前登录用户的关注列表，支持分页和角色筛选
// @Tags 关注
// @Accept json
// @Produce json
// @Param page query int false "页码（从1开始）" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param userRole query int false "过滤用户角色：1=老板,2=陪玩"
// @Success 200 {object} types.GetMyFollowingListResponse "成功"
// @Failure 401 {object} types.BaseResp "未授权"
// @Router /api/user/following [get]
// @Security BearerAuth
func GetMyFollowingListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetMyFollowingListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := follow.NewGetMyFollowingListLogic(r.Context(), svcCtx)
		resp, err := l.GetMyFollowingList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
