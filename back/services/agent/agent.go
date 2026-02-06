package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"SLGaming/back/services/agent/agent"
	"SLGaming/back/services/agent/internal/config"
	agentioc "SLGaming/back/services/agent/internal/ioc"
	"SLGaming/back/services/agent/internal/server"
	"SLGaming/back/services/agent/internal/svc"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

var configFile = flag.String("f", "etc/agent.yaml", "the config file")

func main() {
	flag.Parse()

	var cfg config.Config
	conf.MustLoad(*configFile, &cfg)
	var nacosClient config_client.IConfigClient
	if len(cfg.Nacos.Hosts) > 0 && cfg.Nacos.DataId != "" {
		if client, err := agentioc.InitNacos(cfg.Nacos); err == nil {
			nacosClient = client
			if content, err := agentioc.FetchConfig(client, cfg.Nacos); err == nil && strings.TrimSpace(content) != "" {
				if err := yaml.Unmarshal([]byte(content), &cfg); err == nil {
					logx.Infof("load config from nacos succeeded")
				}
			}
		}
	}

	ctx := svc.NewServiceContext(cfg)

	registrar, err := agentioc.RegisterConsul(cfg.Consul, cfg.ListenOn)
	if err != nil {
		logx.Errorf("consul register failed: %v", err)
	}

	if nacosClient != nil {
		agentioc.ListenConfig(nacosClient, cfg.Nacos, func(content string) {
			if strings.TrimSpace(content) == "" {
				return
			}
			newCfg := ctx.Config()
			if err := yaml.Unmarshal([]byte(content), &newCfg); err == nil {
				ctx.UpdateConfig(newCfg)
			}
		})
	}

	s := zrpc.MustNewServer(cfg.RpcServerConf, func(grpcServer *grpc.Server) {
		agent.RegisterAgentServer(grpcServer, server.NewAgentServer(ctx))

		if cfg.Mode == service.DevMode || cfg.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	// 捕获退出信号，优雅停机
	var stopOnce sync.Once
	stopServer := func() {
		stopOnce.Do(func() {
			logx.Info("shutting down agent rpc server")
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

	//logx.Infof("starting agent rpc server at %s", cfg.ListenOn)
	logx.Infof("Starting rpc server at %s...", cfg.ListenOn)
	s.Start()
}
