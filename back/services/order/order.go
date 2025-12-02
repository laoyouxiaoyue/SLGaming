package main

import (
	"flag"
	"fmt"

	"SLGaming/back/services/order/internal/config"
	orderioc "SLGaming/back/services/order/internal/ioc"
	"SLGaming/back/services/order/internal/server"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/order.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 注册到 Consul（如果配置了）
	var registrar *ioc.ConsulRegistrar
	var err error
	if c.Consul.Address != "" && c.Consul.Service.Name != "" {
		registrar, err = orderioc.RegisterConsul(c.Consul, c.ListenOn)
		if err != nil {
			fmt.Printf("failed to register service to consul: %v\n", err)
		} else {
			defer registrar.Deregister()
		}
	}

	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		order.RegisterOrderServer(grpcServer, server.NewOrderServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
