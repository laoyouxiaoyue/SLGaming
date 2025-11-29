// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"

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

	// 全局应用鉴权中间件（公开接口会在中间件中自动跳过）
	server.Use(middleware.AuthMiddleware(ctx))

	// 注册路由处理器
	handler.RegisterHandlers(server, ctx)

	logx.Infof("Starting gateway server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
