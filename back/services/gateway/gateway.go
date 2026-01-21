// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"net/http"

	"SLGaming/back/services/gateway/internal/handler"
	"SLGaming/back/services/gateway/internal/ioc"
	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()

	// 加载配置（支持从 Nacos 加载）
	c := ioc.LoadConfig(*configFile)

	// 创建 REST 服务器
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 注册 Consul 服务
	registrar, err := ioc.RegisterConsul(c.Consul, fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		logx.Errorf("consul register failed: %v", err)
	} else if registrar != nil {
		defer registrar.Deregister()
	}

	// 创建服务上下文
	ctx := svc.NewServiceContext(c)

	// 全局应用 CORS 中间件（必须在其他中间件之前，使用最宽松的配置）
	server.Use(middleware.CORSMiddleware(nil))

	// 全局应用限流中间件（在鉴权之前，避免无效请求占用资源）
	server.Use(middleware.RateLimitMiddleware(&c.RateLimit))

	// 全局应用鉴权中间件（公开接口会在中间件中自动跳过）
	server.Use(middleware.AuthMiddleware(ctx))

	// 注册路由处理器
	handler.RegisterHandlers(server, ctx)

	// 为所有API路由注册OPTIONS方法支持，确保CORS预检请求能通过
	// 注意：这必须在RegisterHandlers之后
	apiPaths := []string{
		"/api/user/login",
		"/api/user/register",
		"/api/user/logout",
		"/api/user/refresh-token",
		"/api/user/login-by-code",
		"/api/user",
		"/api/user/companion/profile",
		"/api/user/companions",
		"/api/user/forgetPassword",
		"/api/code/send",
		"/api/order",
		"/api/orders",
		"/api/order/accept",
		"/api/order/cancel",
		"/api/order/complete",
		"/api/order/rate",
		"/api/order/start",
	}

	for _, path := range apiPaths {
		server.AddRoute(rest.Route{
			Method: http.MethodOptions,
			Path:   path,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// CORS中间件已经处理了，这里只是确保路由存在
				w.WriteHeader(http.StatusNoContent)
			}),
		})
	}

	logx.Infof("Starting gateway server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
