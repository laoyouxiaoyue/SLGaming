// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout(req *types.LogoutRequest) (resp *types.LogoutResponse, err error) {
	// 从 context 获取用户ID（从 Access Token 中解析的）
	userID, getUserErr := middleware.GetUserID(l.ctx)
	if getUserErr != nil || userID == 0 {
		return &types.LogoutResponse{
			BaseResp: types.BaseResp{
				Code: 401,
				Msg:  "未授权",
			},
		}, nil
	}

	// 撤销用户的所有 Refresh Token（设置用户级别黑名单）
	if l.svcCtx.TokenStore != nil {
		refreshTokenDuration := l.svcCtx.JWT.GetRefreshTokenDuration()
		if err := l.svcCtx.TokenStore.RevokeAllUserTokens(l.ctx, userID, refreshTokenDuration); err != nil {
			l.Errorf("revoke all user tokens failed: %v", err)
			return &types.LogoutResponse{
				BaseResp: types.BaseResp{
					Code: 500,
					Msg:  "撤销 Token 失败",
				},
			}, nil
		}
	}

	return &types.LogoutResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: types.LogoutData{
			Success: true,
		},
	}, nil
}
