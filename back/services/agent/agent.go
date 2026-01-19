package main

import (
	"flag"
	"fmt"
	"strings"

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

	if registrar, err := agentioc.RegisterConsul(cfg.Consul, cfg.ListenOn); err == nil && registrar != nil {
		defer registrar.Deregister()
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
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", cfg.ListenOn)
	s.Start()
}
