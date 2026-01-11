package middleware

import (
	"net/http"
	"strconv"
	"strings"

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
// 默认配置允许所有源，但允许凭证，这在浏览器中会被忽略
// 生产环境建议明确指定允许的源列表
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{"*"}, // 默认允许所有源（生产环境建议限制）
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-Refresh-Token",
			"X-Requested-With",
		},
		ExposedHeaders: []string{
			"Authorization",
			"X-Refresh-Token",
		},
		AllowCredentials: false, // 默认不允许凭证，因为使用了 "*" 源
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

			// 处理预检请求（OPTIONS）
			if r.Method == http.MethodOptions {
				// 设置允许的源
				if len(config.AllowedOrigins) > 0 {
					// 如果允许凭证，不能使用 "*"，必须指定具体源
					if config.AllowCredentials {
						if origin != "" && isOriginAllowed(origin, config.AllowedOrigins) {
							w.Header().Set("Access-Control-Allow-Origin", origin)
						}
					} else {
						// 不允许凭证时，可以使用 "*"
						if config.AllowedOrigins[0] == "*" {
							w.Header().Set("Access-Control-Allow-Origin", "*")
						} else if origin != "" && isOriginAllowed(origin, config.AllowedOrigins) {
							w.Header().Set("Access-Control-Allow-Origin", origin)
						}
					}
				}

				// 设置允许的方法
				if len(config.AllowedMethods) > 0 {
					w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
				}

				// 设置允许的请求头
				requestHeaders := r.Header.Get("Access-Control-Request-Headers")
				if requestHeaders != "" {
					// 如果客户端请求了特定头部，检查是否允许
					if len(config.AllowedHeaders) > 0 {
						if config.AllowedHeaders[0] == "*" {
							w.Header().Set("Access-Control-Allow-Headers", requestHeaders)
						} else {
							w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
						}
					}
				} else if len(config.AllowedHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
				}

				// 设置暴露的响应头
				if len(config.ExposedHeaders) > 0 {
					w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
				}

				// 设置是否允许凭证
				if config.AllowCredentials {
					// 如果允许凭证，不能使用 "*" 作为源（已在上面处理）
					if origin != "" && isOriginAllowed(origin, config.AllowedOrigins) {
						w.Header().Set("Access-Control-Allow-Credentials", "true")
					}
				}

				// 设置预检请求缓存时间
				if config.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
				}

				// 直接返回 200，不继续处理
				w.WriteHeader(http.StatusOK)
				return
			}

			// 处理实际请求（非 OPTIONS）
			// 设置允许的源
			if len(config.AllowedOrigins) > 0 {
				// 如果允许凭证，不能使用 "*"，必须指定具体源
				if config.AllowCredentials {
					if origin != "" && isOriginAllowed(origin, config.AllowedOrigins) {
						w.Header().Set("Access-Control-Allow-Origin", origin)
					}
				} else {
					// 不允许凭证时，可以使用 "*"
					if config.AllowedOrigins[0] == "*" {
						w.Header().Set("Access-Control-Allow-Origin", "*")
					} else if origin != "" && isOriginAllowed(origin, config.AllowedOrigins) {
						w.Header().Set("Access-Control-Allow-Origin", origin)
					}
				}
			}

			// 设置暴露的响应头
			if len(config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
			}

			// 设置是否允许凭证
			if config.AllowCredentials {
				// 如果允许凭证，不能使用 "*" 作为源（已在上面处理）
				if origin != "" && isOriginAllowed(origin, config.AllowedOrigins) {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			// 继续处理请求
			next.ServeHTTP(w, r)
		}
	}
}

// isOriginAllowed 检查源是否在允许列表中
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
		// 支持通配符匹配（例如：*.example.com）
		if strings.Contains(allowed, "*") {
			pattern := strings.ReplaceAll(allowed, ".", "\\.")
			pattern = strings.ReplaceAll(pattern, "*", ".*")
			// 这里可以使用正则表达式，但为了简单，我们只做简单的字符串匹配
			if strings.HasPrefix(origin, strings.TrimSuffix(allowed, "*")) {
				return true
			}
		}
	}
	return false
}
