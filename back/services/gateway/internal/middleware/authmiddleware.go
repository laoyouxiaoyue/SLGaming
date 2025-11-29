package middleware

import (
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
				logx.Errorf("jwt verify failed: %v", err)
				var errMsg string
				if err == jwt.ErrExpiredToken {
					errMsg = "认证令牌已过期"
				} else {
					errMsg = "认证令牌无效"
				}
				httpx.ErrorCtx(r.Context(), w, errors.New(errMsg))
				return
			}

			// 将用户 ID 存储到 context 中
			ctx := SetUserID(r.Context(), claims.UserID)

			// 继续处理请求
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}
