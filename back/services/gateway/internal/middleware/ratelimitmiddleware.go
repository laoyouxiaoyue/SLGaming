package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"SLGaming/back/services/gateway/internal/config"
	"SLGaming/back/services/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	capacity    int64         // 桶容量（令牌数）
	refillRate  int64         // 每秒补充的令牌数
	tokens      int64         // 当前令牌数
	lastRefill  time.Time     // 上次补充令牌的时间
	mu          sync.Mutex    // 互斥锁
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(capacity, refillRate int64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		refillRate: refillRate,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

// Allow 检查是否允许请求（消耗一个令牌）
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	
	// 计算应该补充的令牌数
	refillTokens := int64(elapsed.Seconds()) * tb.refillRate
	if refillTokens > 0 {
		tb.tokens = tb.tokens + refillTokens
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	// 检查是否有可用令牌
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// RateLimiter 限流器管理器
type RateLimiter struct {
	globalLimiter *TokenBucket                    // 全局限流器
	routeLimiters map[string]*TokenBucket        // 路由限流器（key: path+method）
	ipLimiters    map[string]*TokenBucket        // IP 限流器（key: ip）
	userLimiters  map[string]*TokenBucket        // 用户限流器（key: userID）
	routeConfigs  map[string]*config.RouteRateLimitConf // 路由配置（key: path+method）
	mu            sync.RWMutex                    // 读写锁
	cleanupTicker *time.Ticker                    // 清理过期限流器的定时器
}

// NewRateLimiter 创建限流器管理器
func NewRateLimiter(globalQPS int) *RateLimiter {
	rl := &RateLimiter{
		globalLimiter: NewTokenBucket(int64(globalQPS), int64(globalQPS)),
		routeLimiters: make(map[string]*TokenBucket),
		ipLimiters:    make(map[string]*TokenBucket),
		userLimiters:  make(map[string]*TokenBucket),
		routeConfigs:  make(map[string]*config.RouteRateLimitConf),
		cleanupTicker: time.NewTicker(5 * time.Minute), // 每5分钟清理一次
	}

	// 启动清理协程
	go rl.cleanup()

	return rl
}

// cleanup 清理长时间未使用的限流器（防止内存泄漏）
func (rl *RateLimiter) cleanup() {
	for range rl.cleanupTicker.C {
		// 这里可以添加清理逻辑，比如删除超过一定时间未使用的限流器
		// 简化实现：暂时不清理，因为 map 会自动管理
	}
}

// getRouteKey 获取路由 key
func getRouteKey(path, method string) string {
	return method + ":" + path
}

// getRouteLimiter 获取或创建路由限流器
func (rl *RateLimiter) getRouteLimiter(path, method string, qps int) *TokenBucket {
	if qps <= 0 {
		return nil
	}

	key := getRouteKey(path, method)
	rl.mu.RLock()
	limiter, exists := rl.routeLimiters[key]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 双重检查
	if limiter, exists := rl.routeLimiters[key]; exists {
		return limiter
	}

	limiter = NewTokenBucket(int64(qps), int64(qps))
	rl.routeLimiters[key] = limiter
	return limiter
}

// getIPLimiter 获取或创建 IP 限流器
func (rl *RateLimiter) getIPLimiter(ip string, qps int) *TokenBucket {
	if qps <= 0 || ip == "" {
		return nil
	}

	rl.mu.RLock()
	limiter, exists := rl.ipLimiters[ip]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 双重检查
	if limiter, exists := rl.ipLimiters[ip]; exists {
		return limiter
	}

	limiter = NewTokenBucket(int64(qps), int64(qps))
	rl.ipLimiters[ip] = limiter
	return limiter
}

// getUserLimiter 获取或创建用户限流器
func (rl *RateLimiter) getUserLimiter(userID uint64, qps int) *TokenBucket {
	if qps <= 0 || userID == 0 {
		return nil
	}

	// 使用 strconv.FormatUint 将 userID 转为字符串作为 key
	key := "user:" + strconv.FormatUint(userID, 10)
	
	rl.mu.RLock()
	limiter, exists := rl.userLimiters[key]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 双重检查
	if limiter, exists := rl.userLimiters[key]; exists {
		return limiter
	}

	limiter = NewTokenBucket(int64(qps), int64(qps))
	rl.userLimiters[key] = limiter
	return limiter
}

// checkLimit 检查是否超过限流
func (rl *RateLimiter) checkLimit(path, method, ip string, userID uint64, routeConfig *config.RouteRateLimitConf) bool {
	// 1. 检查全局限流
	if !rl.globalLimiter.Allow() {
		logx.Infof("Rate limit exceeded: global limit, path=%s, method=%s", path, method)
		return false
	}

	// 2. 检查路由级别限流
	if routeConfig != nil {
		// 路由全局 QPS
		if routeConfig.GlobalQPS > 0 {
			limiter := rl.getRouteLimiter(path, method, routeConfig.GlobalQPS)
			if limiter != nil && !limiter.Allow() {
				logx.Infof("Rate limit exceeded: route global limit, path=%s, method=%s", path, method)
				return false
			}
		}

		// 每个 IP 的 QPS
		if routeConfig.PerIPQPS > 0 {
			limiter := rl.getIPLimiter(ip, routeConfig.PerIPQPS)
			if limiter != nil && !limiter.Allow() {
				logx.Infof("Rate limit exceeded: per IP QPS limit, path=%s, method=%s, ip=%s", path, method, ip)
				return false
			}
		}

		// 每个用户的 QPS（仅当用户已登录时）
		if routeConfig.PerUserQPS > 0 && userID > 0 {
			limiter := rl.getUserLimiter(userID, routeConfig.PerUserQPS)
			if limiter != nil && !limiter.Allow() {
				logx.Infof("Rate limit exceeded: per user QPS limit, path=%s, method=%s, userID=%d", path, method, userID)
				return false
			}
		}
	}

	return true
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(cfg *config.RateLimitConf) rest.Middleware {
	if !cfg.Enabled {
		// 如果未启用限流，返回空中间件
		return func(next http.HandlerFunc) http.HandlerFunc {
			return next
		}
	}

	limiter := NewRateLimiter(cfg.GlobalQPS)

	// 构建路由配置映射
	routeConfigMap := make(map[string]*config.RouteRateLimitConf)
	for i := range cfg.Routes {
		route := &cfg.Routes[i]
		method := route.Method
		if method == "" {
			method = "*" // 空方法表示匹配所有方法
		}
		key := getRouteKey(route.Path, method)
		routeConfigMap[key] = route
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// OPTIONS 预检请求直接跳过限流
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			path := r.URL.Path
			method := r.Method

			// 获取客户端 IP
			ip := getClientIP(r)

			// 尝试从 context 获取用户 ID（可能未登录）
			userID, _ := GetUserID(r.Context())

			// 查找路由配置
			var routeConfig *config.RouteRateLimitConf
			// 先尝试精确匹配（method + path）
			key := getRouteKey(path, method)
			routeConfig = routeConfigMap[key]
			// 如果没找到，尝试匹配所有方法
			if routeConfig == nil {
				key = getRouteKey(path, "*")
				routeConfig = routeConfigMap[key]
			}

			// 检查限流
			if !limiter.checkLimit(path, method, ip, userID, routeConfig) {
				httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, &types.BaseResp{
					Code: 429,
					Msg:  "请求过于频繁，请稍后再试",
				})
				return
			}

			// 通过限流检查，继续处理请求
			next.ServeHTTP(w, r)
		}
	}
}

// getClientIP 获取客户端真实 IP
func getClientIP(r *http.Request) string {
	// 1. 检查 X-Forwarded-For（代理/负载均衡器）
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For 可能包含多个 IP，取第一个
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 2. 检查 X-Real-IP
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return strings.TrimSpace(ip)
	}

	// 3. 使用 RemoteAddr
	ip = r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}
