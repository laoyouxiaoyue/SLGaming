package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"SLGaming/back/services/user/internal/config"
	"SLGaming/back/services/user/internal/ioc"
	"SLGaming/back/services/user/internal/job"
	"SLGaming/back/services/user/internal/server"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

var configFile = flag.String("f", "etc/user.yaml", "the config file")

func main() {
	flag.Parse()

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

	registrar, err := ioc.RegisterConsul(cfg.Consul, cfg.ListenOn)
	if err != nil {
		logx.Errorf("consul register failed: %v", err)
	} else if registrar != nil {
		defer registrar.Deregister()
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

	// 启动订单退款事件 Consumer 与用户领域事件 Outbox 分发任务
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	job.StartOrderRefundConsumer(rootCtx, ctx)
	job.StartUserOutboxDispatcher(rootCtx, ctx)

	s := zrpc.MustNewServer(cfg.RpcServerConf, func(grpcServer *grpc.Server) {
		user.RegisterUserServer(grpcServer, server.NewUserServer(ctx))

		if cfg.Mode == service.DevMode || cfg.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", cfg.ListenOn)
	s.Start()
}
