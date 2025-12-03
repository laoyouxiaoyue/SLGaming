// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"time"

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
	// 从 context 获取用户ID（由网关鉴权中间件注入）
	userID, _ := middleware.GetUserID(l.ctx)

	if l.svcCtx.TokenStore != nil {
		// 1. 撤销当前的 Access Token
		accessToken, tokenErr := middleware.GetAccessToken(l.ctx)
		if tokenErr == nil && accessToken != "" {
			// 解析 Access Token 获取过期时间
			claims, parseErr := l.svcCtx.JWT.VerifyToken(accessToken)
			if parseErr == nil {
				// 计算剩余有效期
				now := time.Now()
				tokenExpiration := claims.ExpiresAt.Time
				remainingTTL := tokenExpiration.Sub(now)
				if remainingTTL > 0 {
					if err := l.svcCtx.TokenStore.RevokeAccessToken(l.ctx, userID, accessToken, remainingTTL); err != nil {
						l.Errorf("revoke access token failed: %v", err)
						// 继续执行，不因为单个 token 撤销失败而中断
					}
				}
			}
		}

		// 2. 撤销当前的 Refresh Token（如果存在）
		refreshToken, refreshErr := middleware.GetRefreshToken(l.ctx)
		if refreshErr == nil && refreshToken != "" {
			// 解析 Refresh Token 获取过期时间
			claims, parseErr := l.svcCtx.JWT.VerifyToken(refreshToken)
			if parseErr == nil {
				// 计算剩余有效期
				now := time.Now()
				tokenExpiration := claims.ExpiresAt.Time
				remainingTTL := tokenExpiration.Sub(now)
				if remainingTTL > 0 {
					if err := l.svcCtx.TokenStore.RevokeRefreshToken(l.ctx, userID, refreshToken, remainingTTL); err != nil {
						l.Errorf("revoke refresh token failed: %v", err)
						// 继续执行，不因为单个 token 撤销失败而中断
					}
				}
			}
		}
	}

	return &types.LogoutResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "登出成功",
		},
		Data: types.LogoutData{
			Success: true,
		},
	}, nil
}
