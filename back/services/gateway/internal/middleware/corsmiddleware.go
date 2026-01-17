package middleware

import (
	"net/http"
	"regexp"
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

			// 处理预检请求（OPTIONS）
			if r.Method == http.MethodOptions {
				// 设置允许的源（必须设置，即使 origin 为空）
				// 优先处理：如果允许凭证，不能使用 "*"，必须指定具体源
				if config.AllowCredentials {
					if origin != "" && isOriginAllowed(origin, config.AllowedOrigins) {
						w.Header().Set("Access-Control-Allow-Origin", origin)
					} else if len(config.AllowedOrigins) > 0 && !isWildcardOrigin(config.AllowedOrigins) {
						// 如果配置了具体源但没有匹配，不设置（浏览器会拒绝）
						// 但为了兼容性，如果 origin 为空，也设置第一个允许的源
						if origin == "" && len(config.AllowedOrigins) > 0 {
							w.Header().Set("Access-Control-Allow-Origin", config.AllowedOrigins[0])
						}
					}
				} else {
					// 不允许凭证时，可以使用 "*"
					if isWildcardOrigin(config.AllowedOrigins) {
						w.Header().Set("Access-Control-Allow-Origin", "*")
					} else if origin != "" && isOriginAllowed(origin, config.AllowedOrigins) {
						w.Header().Set("Access-Control-Allow-Origin", origin)
					} else {
						// 默认允许所有源（包括 origin 为空的情况）
						w.Header().Set("Access-Control-Allow-Origin", "*")
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
						// 检查是否允许所有请求头
						hasWildcard := false
						for _, header := range config.AllowedHeaders {
							if header == "*" {
								hasWildcard = true
								break
							}
						}
						if hasWildcard {
							// 允许所有请求头，直接返回客户端请求的头部
							w.Header().Set("Access-Control-Allow-Headers", requestHeaders)
						} else {
							// 返回配置的允许头部列表
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
			// 设置允许的源（必须设置）
			// 优先处理：如果允许凭证，不能使用 "*"，必须指定具体源
			if config.AllowCredentials {
				if origin != "" && isOriginAllowed(origin, config.AllowedOrigins) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				} else if len(config.AllowedOrigins) > 0 && !isWildcardOrigin(config.AllowedOrigins) {
					// 如果配置了具体源但没有匹配，不设置（浏览器会拒绝）
					// 但为了兼容性，如果 origin 为空，也设置第一个允许的源
					if origin == "" && len(config.AllowedOrigins) > 0 {
						w.Header().Set("Access-Control-Allow-Origin", config.AllowedOrigins[0])
					}
				}
			} else {
				// 不允许凭证时，可以使用 "*"
				if isWildcardOrigin(config.AllowedOrigins) {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else if origin != "" && isOriginAllowed(origin, config.AllowedOrigins) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				} else {
					// 默认允许所有源（包括 origin 为空的情况）
					w.Header().Set("Access-Control-Allow-Origin", "*")
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

// isWildcardOrigin 检查是否允许所有源（包含 "*"）
func isWildcardOrigin(allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
	}
	return false
}

// isOriginAllowed 检查源是否在允许列表中
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
		// 支持通配符匹配（例如：*.example.com 或 https://*.example.com）
		if strings.Contains(allowed, "*") {
			// 将通配符模式转换为正则表达式
			// 例如：*.example.com -> .*\.example\.com
			// 例如：https://*.example.com -> https://.*\.example\.com
			pattern := strings.ReplaceAll(allowed, ".", "\\.")
			pattern = strings.ReplaceAll(pattern, "*", ".*")
			pattern = "^" + pattern + "$"

			matched, err := regexp.MatchString(pattern, origin)
			if err == nil && matched {
				return true
			}
		}
	}
	return false
}
