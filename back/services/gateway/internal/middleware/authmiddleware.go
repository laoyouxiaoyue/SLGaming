package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"SLGaming/back/services/gateway/internal/jwt"
	"SLGaming/back/services/gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 不需要鉴权的路径白名单
var publicPaths = map[string]bool{
	"/api/code/send":           true,
	"/api/user/register":       true,
	"/api/user/login":          true,
	"/api/user/login-by-code":  true,
	"/api/user/forgetPassword": true,
	"/api/user/refresh-token":  true, // 刷新Token接口（使用RefreshToken，不需要AccessToken）
	"/health":                  true, // 健康检查接口
}

// isPublicPath 检查路径是否需要鉴权
func isPublicPath(path string) bool {
	return publicPaths[path]
}

// AuthMiddleware JWT 鉴权中间件
func AuthMiddleware(svcCtx *svc.ServiceContext) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 检查是否是公开接口，如果是则直接跳过鉴权
			if isPublicPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// 从请求头获取 token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httpx.ErrorCtx(r.Context(), w, errors.New("未提供认证令牌"))
				return
			}

			// 移除 "Bearer " 前缀
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				// 如果没有 Bearer 前缀，尝试直接使用整个值
				tokenString = strings.TrimSpace(authHeader)
			}

			if tokenString == "" {
				httpx.ErrorCtx(r.Context(), w, errors.New("认证令牌格式错误"))
				return
			}

			// 验证 token
			claims, err := svcCtx.JWT.VerifyToken(tokenString)
			if err != nil {
				// 如果 Access Token 过期，尝试使用 Refresh Token 自动刷新
				if err == jwt.ErrExpiredToken {
					// 尝试从请求头获取 Refresh Token
					refreshToken := r.Header.Get("X-Refresh-Token")
					if refreshToken == "" {
						// 如果没有 Refresh Token，返回错误
						httpx.ErrorCtx(r.Context(), w, errors.New("认证令牌已过期"))
						return
					}

					// 尝试自动刷新 Access Token
					newAccessToken, newRefreshToken, userID, refreshErr := tryAutoRefreshToken(r.Context(), svcCtx, refreshToken)
					if refreshErr != nil {
						logx.Errorf("auto refresh token failed: %v", refreshErr)
						httpx.ErrorCtx(r.Context(), w, errors.New("认证令牌已过期，刷新失败"))
						return
					}

					// 将新的 Access Token 设置到响应头
					w.Header().Set("Authorization", "Bearer "+newAccessToken)
					if newRefreshToken != "" {
						w.Header().Set("X-Refresh-Token", newRefreshToken)
					}

					// 将用户 ID 和 Access Token 存储到 context 中
					ctx := SetUserID(r.Context(), userID)
					ctx = SetAccessToken(ctx, newAccessToken)

					// 继续处理请求
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}

				// 其他错误（无效 token）
				logx.Errorf("jwt verify failed: %v", err)
				httpx.ErrorCtx(r.Context(), w, errors.New("认证令牌无效"))
				return
			}

			// 验证 Access Token 是否在黑名单中（已被撤销）
			if svcCtx.TokenStore != nil {
				tokenExpiration := claims.ExpiresAt.Time
				valid, err := svcCtx.TokenStore.VerifyAccessToken(r.Context(), claims.UserID, tokenString, tokenExpiration)
				if err != nil {
					logx.Errorf("verify access token failed: %v", err)
					httpx.ErrorCtx(r.Context(), w, errors.New("验证认证令牌失败"))
					return
				}
				if !valid {
					httpx.ErrorCtx(r.Context(), w, errors.New("认证令牌已被撤销"))
					return
				}
			}

			// 将用户 ID 和 Access Token 存储到 context 中
			ctx := SetUserID(r.Context(), claims.UserID)
			ctx = SetAccessToken(ctx, tokenString)

			// 继续处理请求
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

// tryAutoRefreshToken 尝试使用 Refresh Token 自动刷新 Access Token
// 返回: newAccessToken, refreshToken(保持不变), userID, error
func tryAutoRefreshToken(ctx context.Context, svcCtx *svc.ServiceContext, refreshToken string) (string, string, uint64, error) {
	// 验证 Refresh Token
	claims, err := svcCtx.JWT.VerifyToken(refreshToken)
	if err != nil {
		if err == jwt.ErrExpiredToken {
			return "", "", 0, errors.New("Refresh Token 已过期")
		}
		return "", "", 0, errors.New("Refresh Token 无效")
	}

	// 验证 Refresh Token 是否在黑名单中（未被撤销）
	if svcCtx.TokenStore != nil {
		tokenExpiration := claims.ExpiresAt.Time
		valid, err := svcCtx.TokenStore.VerifyRefreshToken(ctx, claims.UserID, refreshToken, tokenExpiration)
		if err != nil {
			logx.Errorf("verify refresh token failed: %v", err)
			return "", "", 0, errors.New("验证 Refresh Token 失败")
		}
		if !valid {
			return "", "", 0, errors.New("Refresh Token 已被撤销")
		}
	}

	// 只生成新的 Access Token，Refresh Token 保持不变
	accessToken, err := svcCtx.JWT.GenerateAccessToken(claims.UserID)
	if err != nil {
		logx.Errorf("generate access token failed: %v", err)
		return "", "", 0, errors.New("生成 Access Token 失败")
	}

	// Refresh Token 不刷新，保持原样返回
	return accessToken, refreshToken, claims.UserID, nil
}
