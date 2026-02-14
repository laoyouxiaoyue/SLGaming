package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"SLGaming/back/pkg/ioc"
	"SLGaming/back/services/order/internal/config"
	"SLGaming/back/services/order/internal/helper"
	orderioc "SLGaming/back/services/order/internal/ioc"
	"SLGaming/back/services/order/internal/job"
	_ "SLGaming/back/services/order/internal/metrics"
	"SLGaming/back/services/order/internal/server"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/order.yaml", "the config file")

func getAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()
	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

func startMetricsServer(preferredPort int) (int, error) {
	port := preferredPort
	if port <= 0 {
		var err error
		port, err = getAvailablePort()
		if err != nil {
			return 0, fmt.Errorf("failed to get available port: %w", err)
		}
		helper.LogInfo(logx.WithContext(context.Background()), helper.OpServer, "auto-assigned metrics port", map[string]interface{}{
			"port": port,
		})
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		addr := fmt.Sprintf(":%d", port)
		helper.LogInfo(logx.WithContext(context.Background()), helper.OpServer, "metrics server started", map[string]interface{}{
			"listen_on": addr,
		})
		if err := http.ListenAndServe(addr, nil); err != nil {
			helper.LogError(logx.WithContext(context.Background()), helper.OpServer, "metrics server failed", err, map[string]interface{}{
				"listen_on": addr,
			})
		}
	}()

	return port, nil
}

func main() {
	flag.Parse()

	logger := logx.WithContext(context.Background())

	var c config.Config
	conf.MustLoad(*configFile, &c)

	metricsPort, err := startMetricsServer(c.MetricsPort)
	if err != nil {
		helper.LogError(logger, helper.OpServer, "start metrics server failed", err, nil)
	}

	var registrar *ioc.ConsulRegistrar
	if c.Consul.Address != "" && c.Consul.Service.Name != "" {
		registrar, err = orderioc.RegisterConsul(c.Consul, c.ListenOn, metricsPort)
		if err != nil {
			helper.LogError(logger, helper.OpServer, "consul register failed", err, map[string]interface{}{
				"listen_on": c.ListenOn,
			})
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
			helper.LogInfo(logger, helper.OpServer, "shutting down", map[string]interface{}{
				"reason": "signal",
			})
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

	helper.LogSuccess(logger, helper.OpServer, map[string]interface{}{
		"listen_on":    c.ListenOn,
		"metrics_port": metricsPort,
		"mode":         c.Mode,
	})

	s.Start()
}
