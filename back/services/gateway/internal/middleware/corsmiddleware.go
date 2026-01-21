package middleware

import (
	"fmt"
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowedOrigins   []string // 允许的源（例如：["http://localhost:3000", "https://example.com"]）
	AllowedMethods   []string // 允许的 HTTP 方法（例如：["GET", "POST", "PUT", "DELETE", "OPTIONS"]）
	AllowedHeaders   []string // 允许的请求头（例如：["Content-Type", "Authorization", "X-Refresh-Token"]）
	ExposedHeaders   []string // 暴露给客户端的响应头
	AllowCredentials bool     // 是否允许发送凭证（cookies、authorization headers 等）
	MaxAge           int      // 预检请求的缓存时间（秒）
}

// DefaultCORSConfig 返回默认的 CORS 配置
// 注意：如果 AllowCredentials 为 true，AllowedOrigins 不能包含 "*"
// 默认配置允许所有源、所有方法、所有请求头（最宽松配置）
// 生产环境建议明确指定允许的源列表
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{"*"}, // 允许所有源
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowedHeaders: []string{"*"}, // 允许所有请求头
		ExposedHeaders: []string{
			"Authorization",
			"X-Refresh-Token",
			"Content-Type",
		},
		AllowCredentials: false, // 不允许凭证，因为使用了 "*" 源
		MaxAge:           86400, // 24 小时
	}
}

// CORSMiddleware CORS 跨域中间件
func CORSMiddleware(config *CORSConfig) rest.Middleware {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			fmt.Println("CORS Middleware - Origin:", origin, "Method:", r.Method, "Path:", r.URL.Path)

			// 允许所有源访问
			// 如果设置了 credentials，必须返回具体的 origin（不能是 "*"）
			// 如果没有 origin，使用 "*"
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			// 允许所有方法和请求头
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Expose-Headers", "Authorization, X-Refresh-Token, Content-Type")

			// 如果请求有 origin，允许 credentials
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			w.Header().Set("Access-Control-Max-Age", "86400")

			// 预检请求直接返回 204，不继续处理
			if r.Method == http.MethodOptions {
				fmt.Println("CORS: Handling OPTIONS request, returning 204")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// 继续处理实际请求
			next.ServeHTTP(w, r)
		}
	}
}
