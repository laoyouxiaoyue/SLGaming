package user

import (
	"context"

	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// generateTokens 生成 Access Token 和 Refresh Token
func generateTokens(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64, logger logx.Logger) (*types.LoginData, error) {
	// 生成 Access Token
	accessToken, err := svcCtx.JWT.GenerateAccessToken(userID)
	if err != nil {
		logger.Errorf("generate access token failed: %v", err)
		return nil, err
	}

	// 生成 Refresh Token
	refreshToken, err := svcCtx.JWT.GenerateRefreshToken(userID)
	if err != nil {
		logger.Errorf("generate refresh token failed: %v", err)
		return nil, err
	}

	// 存储 Refresh Token
	if svcCtx.TokenStore != nil {
		refreshTokenDuration := svcCtx.JWT.GetRefreshTokenDuration()
		if err := svcCtx.TokenStore.StoreRefreshToken(ctx, userID, refreshToken, refreshTokenDuration); err != nil {
			logger.Errorf("store refresh token failed: %v", err)
			// 不返回错误，只记录日志
		}
	}

	return &types.LoginData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(svcCtx.JWT.GetAccessTokenDuration().Seconds()),
	}, nil
}
