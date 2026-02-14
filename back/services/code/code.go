package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"SLGaming/back/services/code/code"
	"SLGaming/back/services/code/internal/ioc"
	"SLGaming/back/services/code/internal/server"
	"SLGaming/back/services/code/internal/svc"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	configFile    = flag.String("f", "etc/code.yaml", "the config file")
	templatesFile = flag.String("t", "etc/code-templates.yaml", "the templates file")
)

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
		logx.Infof("[server] info: auto-assigned metrics port, port=%d", port)
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		addr := fmt.Sprintf(":%d", port)
		logx.Infof("[server] info: metrics server started, listen_on=%s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			logx.Errorf("[server] failed: metrics server failed, listen_on=%s, error=%v", addr, err)
		}
	}()

	return port, nil
}

func main() {
	flag.Parse()

	c := ioc.LoadConfig(*configFile, *templatesFile)
	ctx := svc.NewServiceContext(c)

	metricsPort, err := startMetricsServer(c.MetricsPort)
	if err != nil {
		logx.Errorf("[server] failed: start metrics server failed, error=%v", err)
	}

	registrar, err := ioc.RegisterConsul(c.Consul, c.ListenOn, metricsPort)
	if err != nil {
		logx.Errorf("[server] failed: consul register failed, listen_on=%s, error=%v", c.ListenOn, err)
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		code.RegisterCodeServer(grpcServer, server.NewCodeServer(ctx))
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	var stopOnce sync.Once
	stopServer := func() {
		stopOnce.Do(func() {
			logx.Infof("[server] info: shutting down, reason=signal")
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

	logx.Infof("[server] succeeded: service started, listen_on=%s, metrics_port=%d, mode=%s", c.ListenOn, metricsPort, c.Mode)

	s.Start()
}
