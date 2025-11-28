// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"fmt"

	"SLGaming/back/services/code/codeclient"
	"SLGaming/back/services/gateway/internal/config"
	"SLGaming/back/services/user/userclient"

	"github.com/hashicorp/consul/api"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	CodeRPC codeclient.Code
	UserRPC userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	ctx := &ServiceContext{
		Config: c,
	}

	if c.Upstream.CodeService != "" {
		if cli, err := newRPCClient(c.Consul, c.Upstream.CodeService); err != nil {
			logx.Errorf("init code rpc client failed: %v", err)
		} else {
			ctx.CodeRPC = codeclient.NewCode(cli)
		}
	}

	if c.Upstream.UserService != "" {
		if cli, err := newRPCClient(c.Consul, c.Upstream.UserService); err != nil {
			logx.Errorf("init user rpc client failed: %v", err)
		} else {
			ctx.UserRPC = userclient.NewUser(cli)
		}
	}

	return ctx
}

func newRPCClient(consulConf config.ConsulConf, serviceName string) (zrpc.Client, error) {
	endpoints, err := resolveServiceEndpoints(consulConf, serviceName)
	if err != nil {
		return nil, err
	}
	return zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: endpoints,
		NonBlock:  true,
	}), nil
}

func resolveServiceEndpoints(conf config.ConsulConf, serviceName string) ([]string, error) {
	if conf.Address == "" {
		return nil, fmt.Errorf("consul address is empty")
	}
	if serviceName == "" {
		return nil, fmt.Errorf("service name is empty")
	}

	client, err := api.NewClient(&api.Config{
		Address: conf.Address,
		Token:   conf.Token,
	})
	if err != nil {
		return nil, fmt.Errorf("create consul client: %w", err)
	}

	entries, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("query service %s: %w", serviceName, err)
	}

	var endpoints []string
	for _, entry := range entries {
		addr := entry.Service.Address
		if addr == "" {
			addr = entry.Node.Address
		}
		if addr == "" {
			continue
		}
		endpoints = append(endpoints, fmt.Sprintf("%s:%d", addr, entry.Service.Port))
	}
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no healthy instances for %s", serviceName)
	}
	return endpoints, nil
}
