package middleware

import (
	"context"
	"errors"
	"net/http"
	"path"
	"strings"

	"SLGaming/back/services/gateway/internal/jwt"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"

	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 不需要鉴权的路径白名单
var publicPaths = map[string]bool{
	"/api/code/send":                     true,
	"/api/user/register":                 true,
	"/api/user/login":                    true,
	"/api/user/login-by-code":            true,
	"/api/user/forgetPassword":           true,
	"/api/user/refresh-token":            true, // 刷新Token接口（使用RefreshToken，不需要AccessToken）
	"/api/user/companions":               true, // 获取陪玩列表
	"/api/user/companion/profile/public": true, // 公开获取陪玩信息
	"/api/user/gameskills":               true, // 获取游戏技能列表
	"/health":                            true, // 健康检查接口
}

// isPublicPath 检查路径是否需要鉴权
func isPublicPath(path string) bool {
	if publicPaths[path] {
		return true
	}
	cleaned := pathClean(path)
	return publicPaths[cleaned]
}

func pathClean(p string) string {
	if p == "" {
		return "/"
	}
	cleaned := path.Clean(p)
	// 保留原始结尾斜杠的兼容性：/api/user/gameskills/ -> /api/user/gameskills
	if strings.HasSuffix(p, "/") && cleaned != "/" {
		cleaned = strings.TrimSuffix(cleaned, "/")
	}
	return cleaned
}

// AuthMiddleware JWT 鉴权中间件
func AuthMiddleware(svcCtx *svc.ServiceContext) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// OPTIONS 预检请求直接跳过鉴权（CORS 中间件已处理）
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// 检查是否是公开接口，如果是则直接跳过鉴权
			if isPublicPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// 从请求头获取 token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, &types.BaseResp{
					Code: 401,
					Msg:  "未提供认证令牌",
				})
				return
			}

			// 移除 "Bearer " 前缀
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				// 如果没有 Bearer 前缀，尝试直接使用整个值
				tokenString = strings.TrimSpace(authHeader)
			}

			if tokenString == "" {
				httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, &types.BaseResp{
					Code: 401,
					Msg:  "认证令牌格式错误",
				})
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
						httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, &types.BaseResp{
							Code: 401,
							Msg:  "认证令牌已过期，请重新登录",
						})
						return
					}

					// 尝试自动刷新 Access Token
					newAccessToken, newRefreshToken, userID, refreshErr := tryAutoRefreshToken(r.Context(), svcCtx, refreshToken)
					if refreshErr != nil {
						logx.Errorf("auto refresh token failed: %v", refreshErr)
						httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, &types.BaseResp{
							Code: 401,
							Msg:  "认证令牌已过期，刷新失败，请重新登录",
						})
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

				// 其他错误：根据错误类型返回更详细的错误信息
				logx.Errorf("jwt verify failed: %v", err)
				errMsg := getTokenErrorMessage(err)
				httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, &types.BaseResp{
					Code: 401,
					Msg:  errMsg,
				})
				return
			}

			// 验证 Access Token 是否在黑名单中（已被撤销）
			if svcCtx.TokenStore != nil {
				tokenExpiration := claims.ExpiresAt.Time
				valid, err := svcCtx.TokenStore.VerifyAccessToken(r.Context(), claims.UserID, tokenString, tokenExpiration)
				if err != nil {
					logx.Errorf("verify access token failed: %v", err)
					httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, &types.BaseResp{
						Code: 401,
						Msg:  "验证认证令牌失败",
					})
					return
				}
				if !valid {
					httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, &types.BaseResp{
						Code: 401,
						Msg:  "认证令牌已被撤销",
					})
					return
				}
			}

			// 将用户 ID 和 Access Token 存储到 context 中
			ctx := SetUserID(r.Context(), claims.UserID)
			ctx = SetAccessToken(ctx, tokenString)

			// 调试日志：记录提取的用户 ID
			logx.Infof("JWT middleware: extracted user_id=%d from token for path=%s", claims.UserID, r.URL.Path)

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

// getTokenErrorMessage 根据 JWT 验证错误返回用户友好的错误信息
func getTokenErrorMessage(err error) string {
	if err == nil {
		return "认证令牌验证失败"
	}

	// 检查是否是 JWT ValidationError
	if ve, ok := err.(*jwtv4.ValidationError); ok {
		if ve.Errors&jwtv4.ValidationErrorMalformed != 0 {
			return "认证令牌格式错误，请检查令牌是否正确"
		}
		if ve.Errors&jwtv4.ValidationErrorUnverifiable != 0 {
			return "认证令牌无法验证，可能是签名算法不匹配"
		}
		if ve.Errors&jwtv4.ValidationErrorSignatureInvalid != 0 {
			return "认证令牌签名无效，令牌可能被篡改"
		}
		if ve.Errors&jwtv4.ValidationErrorNotValidYet != 0 {
			return "认证令牌尚未生效，请稍后再试"
		}
		if ve.Errors&jwtv4.ValidationErrorExpired != 0 {
			return "认证令牌已过期，请重新登录"
		}
		// 其他验证错误
		return "认证令牌验证失败，请重新登录"
	}

	// 如果是自定义错误
	if err == jwt.ErrInvalidToken {
		return "认证令牌无效，请重新登录"
	}

	// 默认错误信息
	return "认证令牌验证失败，请重新登录"
}
