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

// getAvailablePort 获取一个可用的随机端口
func getAvailablePort() (int, error) {
	// 监听随机端口（:0 表示让系统自动分配）
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	// 获取分配的端口
	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// startMetricsServer 启动 metrics HTTP 服务，返回实际使用的端口
func startMetricsServer(preferredPort int) (int, error) {
	port := preferredPort

	// 如果配置为 0 或负数，则使用随机端口
	if port <= 0 {
		var err error
		port, err = getAvailablePort()
		if err != nil {
			return 0, fmt.Errorf("failed to get available port: %w", err)
		}
		logx.Infof("Auto-assigned metrics port: %d", port)
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		addr := fmt.Sprintf(":%d", port)
		logx.Infof("Starting metrics server at %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			logx.Errorf("metrics server failed: %v", err)
		}
	}()

	return port, nil
}

func main() {
	flag.Parse()

	c := ioc.LoadConfig(*configFile, *templatesFile)
	ctx := svc.NewServiceContext(c)

	// 启动 Prometheus metrics HTTP 服务（自动分配端口）
	metricsPort, err := startMetricsServer(c.MetricsPort)
	if err != nil {
		logx.Errorf("failed to start metrics server: %v", err)
		// 不退出，继续启动 gRPC 服务
	}

	registrar, err := ioc.RegisterConsul(c.Consul, c.ListenOn, metricsPort)
	if err != nil {
		logx.Errorf("consul register failed: %v", err)
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		code.RegisterCodeServer(grpcServer, server.NewCodeServer(ctx))
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	// 捕获退出信号，优雅停机
	var stopOnce sync.Once
	stopServer := func() {
		stopOnce.Do(func() {
			logx.Info("shutting down code rpc server")
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

	logx.Info(c.Template)
	fmt.Printf("Starting rpc server at %s, metrics at :%d\n", c.ListenOn, metricsPort)
	s.Start()
}
