package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"SLGaming/back/pkg/ioc"
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
		}
	}

	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		order.RegisterOrderServer(grpcServer, server.NewOrderServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	// 捕获退出信号，优雅停机
	var stopOnce sync.Once
	stopServer := func() {
		stopOnce.Do(func() {
			fmt.Println("shutting down order rpc server")
			s.Stop()
			if registrar != nil {
				registrar.Deregister()
			}
		})
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		stopServer()
	}()

	defer stopServer()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
