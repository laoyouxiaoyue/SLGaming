package user

import (
	"context"

	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

// generateTokens 生成 Access Token 和 Refresh Token
func generateTokens(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64, logger logx.Logger) (*types.LoginData, error) {
	role := int32(0)
	if svcCtx.UserRPC != nil {
		roleResp, err := svcCtx.UserRPC.GetUser(ctx, &userclient.GetUserRequest{Id: userID})
		if err == nil && roleResp != nil && roleResp.User != nil {
			role = roleResp.User.Role
		}
	}

	// 生成 Access Token
	accessToken, err := svcCtx.JWT.GenerateAccessToken(userID, role)
	if err != nil {
		logger.Errorf("generate access token failed: %v", err)
		return nil, err
	}

	// 生成 Refresh Token
	refreshToken, err := svcCtx.JWT.GenerateRefreshToken(userID, role)
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
