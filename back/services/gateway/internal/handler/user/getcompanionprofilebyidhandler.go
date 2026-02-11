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

// GetCompanionProfileByIdHandler 获取指定陪玩资料
// @Summary 获取指定陪玩资料
// @Description 根据用户ID获取指定陪玩的公开资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Param userId query uint64 true "用户ID"
// @Success 200 {object} types.GetCompanionProfileResponse "成功"
// @Failure 400 {object} types.BaseResp "请求参数错误"
// @Router /api/user/companion/profile/public [get]
func GetCompanionProfileByIdHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetCompanionProfileByIdRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewGetCompanionProfileByIdLogic(r.Context(), svcCtx)
		resp, err := l.GetCompanionProfileById(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			utils.WriteResponse(r.Context(), w, resp)
		}
	}
}
