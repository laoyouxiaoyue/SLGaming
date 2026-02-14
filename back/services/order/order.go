package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"SLGaming/back/pkg/ioc"
	"SLGaming/back/services/order/internal/config"
	orderioc "SLGaming/back/services/order/internal/ioc"
	"SLGaming/back/services/order/internal/job"
	"SLGaming/back/services/order/internal/server"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
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

	var registrar *ioc.ConsulRegistrar
	var err error
	if c.Consul.Address != "" && c.Consul.Service.Name != "" {
		registrar, err = orderioc.RegisterConsul(c.Consul, c.ListenOn)
		if err != nil {
			logx.Errorf("[server] failed: consul register failed, listen_on=%s, error=%v", c.ListenOn, err)
		}
	}

	ctx := svc.NewServiceContext(c)

	consumerCtx, cancel := context.WithCancel(context.Background())
	job.StartPaymentStatusConsumer(consumerCtx, ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		order.RegisterOrderServer(grpcServer, server.NewOrderServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	var stopOnce sync.Once
	stopServer := func() {
		stopOnce.Do(func() {
			logx.Infof("[server] info: shutting down, reason=signal")
			cancel()
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

	logx.Infof("[server] succeeded: service started, listen_on=%s, mode=%s", c.ListenOn, c.Mode)

	s.Start()
}
