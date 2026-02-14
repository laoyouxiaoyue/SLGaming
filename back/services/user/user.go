package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"SLGaming/back/services/user/internal/config"
	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/ioc"
	"SLGaming/back/services/user/internal/job"
	_ "SLGaming/back/services/user/internal/metrics"
	"SLGaming/back/services/user/internal/server"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

var configFile = flag.String("f", "etc/user.yaml", "the config file")

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

	var cfg config.Config
	conf.MustLoad(*configFile, &cfg)

	logger := logx.WithContext(context.Background())

	var (
		nacosClient config_client.IConfigClient
		err         error
	)

	if len(cfg.Nacos.Hosts) > 0 && cfg.Nacos.DataId != "" {
		nacosClient, err = ioc.InitNacos(cfg.Nacos)
		if err != nil {
			helper.LogError(logger, helper.OpServer, "init nacos failed, using local config", err, nil)
		} else {
			content, err := ioc.FetchConfig(nacosClient, cfg.Nacos)
			if err != nil {
				helper.LogError(logger, helper.OpServer, "fetch config from nacos failed, using local config", err, nil)
			} else if strings.TrimSpace(content) != "" {
				remoteCfg := cfg
				if err := yaml.Unmarshal([]byte(content), &remoteCfg); err != nil {
					helper.LogError(logger, helper.OpServer, "unmarshal nacos config failed, using local config", err, nil)
				} else {
					cfg = remoteCfg
					helper.LogInfo(logger, helper.OpServer, "load config from nacos succeeded", nil)
				}
			}
		}
	}

	ctx := svc.NewServiceContext(cfg)

	metricsPort, err := startMetricsServer(cfg.MetricsPort)
	if err != nil {
		helper.LogError(logger, helper.OpServer, "start metrics server failed", err, nil)
	}

	registrar, err := ioc.RegisterConsul(cfg.Consul, cfg.ListenOn, metricsPort)
	if err != nil {
		helper.LogError(logger, helper.OpServer, "consul register failed", err, map[string]interface{}{
			"listen_on": cfg.ListenOn,
		})
	}

	if nacosClient != nil {
		if err := ioc.ListenConfig(nacosClient, cfg.Nacos, func(content string) {
			if strings.TrimSpace(content) == "" {
				helper.LogInfo(logger, helper.OpServer, "nacos update skipped: empty content", nil)
				return
			}
			newCfg := ctx.Config()
			if err := yaml.Unmarshal([]byte(content), &newCfg); err != nil {
				helper.LogError(logger, helper.OpServer, "unmarshal nacos config on update failed", err, nil)
				return
			}
			if err := ctx.UpdateConfig(newCfg); err != nil {
				helper.LogError(logger, helper.OpServer, "update service context config failed", err, nil)
				return
			}
			helper.LogInfo(logger, helper.OpServer, "service config hot updated from nacos", nil)
		}); err != nil {
			helper.LogError(logger, helper.OpServer, "listen nacos config failed", err, nil)
		}
	}

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	job.StartOrderRefundConsumer(rootCtx, ctx)
	job.StartRechargeEventConsumer(rootCtx, ctx)
	job.StartAvatarModerationConsumer(rootCtx, ctx)
	job.StartFollowEventConsumer(rootCtx, ctx)

	helper.WarmupRankingFromMySQLAsync(ctx, logx.WithContext(rootCtx))

	s := zrpc.MustNewServer(cfg.RpcServerConf, func(grpcServer *grpc.Server) {
		user.RegisterUserServer(grpcServer, server.NewUserServer(ctx))

		if cfg.Mode == service.DevMode || cfg.Mode == service.TestMode {
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
		"listen_on":    cfg.ListenOn,
		"metrics_port": metricsPort,
		"mode":         cfg.Mode,
	})

	s.Start()
}
