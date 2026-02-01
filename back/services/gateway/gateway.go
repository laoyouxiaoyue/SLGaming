// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	pkgIoc "SLGaming/back/pkg/ioc"
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

	// Consul 注册器（用于退出时注销）
	var registrar *pkgIoc.ConsulRegistrar

	// 优雅停机：确保信号时只 Stop/Deregister 一次
	var stopOnce sync.Once
	stopServer := func() {
		stopOnce.Do(func() {
			logx.Info("shutting down gateway server")
			server.Stop()
			if registrar != nil {
				registrar.Deregister()
			}
		})
	}

	// 注册 Consul 服务
	var err error
	registrar, err = ioc.RegisterConsul(c.Consul, fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		logx.Errorf("consul register failed: %v", err)
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

	// 静态资源服务
	// 映射 /uploads/ 路径到当前运行目录下的 uploads 文件夹
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/uploads/:file",
		Handler: http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))).ServeHTTP,
	})

	// 为所有 /api/* 路径自动添加 OPTIONS 方法支持，确保 CORS 预检请求能通过
	// 使用路径参数匹配所有可能的路径层级
	// 注意：go-zero 的路由匹配是精确匹配，所以需要添加多个通配符路由来覆盖所有情况
	apiOptionsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS 中间件已经处理了 CORS 头，这里只是确保路由存在并返回 204
		w.WriteHeader(http.StatusNoContent)
	})

	// 支持单级路径：/api/:path
	server.AddRoute(rest.Route{
		Method:  http.MethodOptions,
		Path:    "/api/:path",
		Handler: apiOptionsHandler,
	})

	// 支持两级路径：/api/:path1/:path2
	server.AddRoute(rest.Route{
		Method:  http.MethodOptions,
		Path:    "/api/:path1/:path2",
		Handler: apiOptionsHandler,
	})

	// 支持三级路径：/api/:path1/:path2/:path3
	server.AddRoute(rest.Route{
		Method:  http.MethodOptions,
		Path:    "/api/:path1/:path2/:path3",
		Handler: apiOptionsHandler,
	})

	// 支持四级路径：/api/:path1/:path2/:path3/:path4
	server.AddRoute(rest.Route{
		Method:  http.MethodOptions,
		Path:    "/api/:path1/:path2/:path3/:path4",
		Handler: apiOptionsHandler,
	})

	// 也支持 /api 路径本身
	server.AddRoute(rest.Route{
		Method:  http.MethodOptions,
		Path:    "/api",
		Handler: apiOptionsHandler,
	})

	// 捕获退出信号，触发优雅停机
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		stopServer()
	}()

	defer stopServer()

	logx.Infof("Starting gateway server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
