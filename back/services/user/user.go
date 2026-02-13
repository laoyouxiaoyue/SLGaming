package main

import (
	"SLGaming/back/services/user/internal/config"
	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/ioc"
	"SLGaming/back/services/user/internal/job"
	_ "SLGaming/back/services/user/internal/metrics"
	"SLGaming/back/services/user/internal/server"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"
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

	// 配置日志

	// 先加载本地配置
	var cfg config.Config
	conf.MustLoad(*configFile, &cfg)

	var (
		nacosClient config_client.IConfigClient
		err         error
	)

	// 如果配置了 Nacos，尝试从 Nacos 获取配置
	if len(cfg.Nacos.Hosts) > 0 && cfg.Nacos.DataId != "" {
		nacosClient, err = ioc.InitNacos(cfg.Nacos)
		if err != nil {
			logx.Errorf("init nacos failed, using local config: %v", err)
		} else {
			content, err := ioc.FetchConfig(nacosClient, cfg.Nacos)
			if err != nil {
				logx.Errorf("fetch config from nacos failed, using local config: %v", err)
			} else if strings.TrimSpace(content) != "" {
				remoteCfg := cfg
				if err := yaml.Unmarshal([]byte(content), &remoteCfg); err != nil {
					logx.Errorf("unmarshal nacos config failed, using local config: %v", err)
				} else {
					cfg = remoteCfg
					logx.Infof("load config from nacos succeeded")
				}
			}
		}
	}

	ctx := svc.NewServiceContext(cfg)

	metricsPort, err := startMetricsServer(cfg.MetricsPort)
	if err != nil {
		logx.Errorf("failed to start metrics server: %v", err)
	}

	registrar, err := ioc.RegisterConsul(cfg.Consul, cfg.ListenOn, metricsPort)
	if err != nil {
		logx.Errorf("consul register failed: %v", err)
	}

	if nacosClient != nil {
		if err := ioc.ListenConfig(nacosClient, cfg.Nacos, func(content string) {
			if strings.TrimSpace(content) == "" {
				logx.Infof("nacos update skipped: empty content")
				return
			}
			newCfg := ctx.Config()
			if err := yaml.Unmarshal([]byte(content), &newCfg); err != nil {
				logx.Errorf("unmarshal nacos config on update failed: %v", err)
				return
			}
			if err := ctx.UpdateConfig(newCfg); err != nil {
				logx.Errorf("update service context config failed: %v", err)
				return
			}
			logx.Infof("service config hot updated from nacos")
		}); err != nil {
			logx.Errorf("listen nacos config failed: %v", err)
		}
	}

	// 启动订单退款事件 Consumer
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	job.StartOrderRefundConsumer(rootCtx, ctx)
	job.StartRechargeEventConsumer(rootCtx, ctx)
	job.StartAvatarModerationConsumer(rootCtx, ctx)
	job.StartFollowEventConsumer(rootCtx, ctx)

	// 注意：布隆过滤器数据已持久化在 Redis 中（RDB/AOF），正常重启不需要重新加载
	// 只有在以下情况才需要手动初始化：
	// 1. 首次部署（Redis 中还没有布隆过滤器数据）
	// 2. Redis 数据被清空
	// 3. 需要重建布隆过滤器
	// 如需初始化，请使用命令行工具或管理接口执行，不要在启动时全量加载

	// 排行榜异步预热（从MySQL加载数据到Redis，不阻塞启动）
	helper.WarmupRankingFromMySQLAsync(ctx, logx.WithContext(rootCtx))

	s := zrpc.MustNewServer(cfg.RpcServerConf, func(grpcServer *grpc.Server) {
		user.RegisterUserServer(grpcServer, server.NewUserServer(ctx))

		if cfg.Mode == service.DevMode || cfg.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	// 捕获退出信号，优雅停机
	var stopOnce sync.Once
	stopServer := func() {
		stopOnce.Do(func() {
			logx.Info("shutting down user rpc server")
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

	fmt.Printf("Starting rpc server at %s, metrics at :%d\n", cfg.ListenOn, metricsPort)
	s.Start()
}
