package ioc

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"SLGaming/back/services/gateway/internal/config"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"github.com/zeromicro/go-zero/core/logx"
)

type consulRegistrar struct {
	client    *api.Client
	serviceID string
}

// RegisterConsul 注册服务到 Consul
func RegisterConsul(cfg config.ConsulConf, listenOn string) (*consulRegistrar, error) {
	if cfg.Address == "" || cfg.Service.Name == "" {
		return nil, nil
	}

	host, portStr, err := net.SplitHostPort(listenOn)
	if err != nil {
		return nil, fmt.Errorf("split listen address: %w", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid listen port: %w", err)
	}

	serviceAddr := cfg.Service.Address
	if strings.TrimSpace(serviceAddr) == "" || serviceAddr == "0.0.0.0" {
		serviceAddr = host
	}

	serviceID := cfg.Service.ID
	if serviceID == "" {
		serviceID = fmt.Sprintf("%s-%s", cfg.Service.Name, uuid.NewString())
	}

	client, err := api.NewClient(&api.Config{
		Address: cfg.Address,
		Token:   cfg.Token,
	})
	if err != nil {
		return nil, fmt.Errorf("create consul client: %w", err)
	}

	checkInterval := cfg.Service.CheckInterval
	if checkInterval == "" {
		checkInterval = "10s"
	}
	checkTimeout := cfg.Service.CheckTimeout
	if checkTimeout == "" {
		checkTimeout = "5s"
	}

	reg := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    cfg.Service.Name,
		Address: serviceAddr,
		Port:    port,
		Tags:    cfg.Service.Tags,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", serviceAddr, port),
			Interval: checkInterval,
			Timeout:  checkTimeout,
		},
	}

	if err := client.Agent().ServiceRegister(reg); err != nil {
		return nil, fmt.Errorf("register consul service: %w", err)
	}
	logx.Infof("registered consul service %s (%s:%d)", reg.Name, serviceAddr, port)

	return &consulRegistrar{
		client:    client,
		serviceID: serviceID,
	}, nil
}

func (r *consulRegistrar) Deregister() {
	if r == nil {
		return
	}
	if err := r.client.Agent().ServiceDeregister(r.serviceID); err != nil {
		logx.Errorf("deregister consul service failed: %v", err)
	}
}

// ResolveServiceEndpoints 通过 Consul 服务发现解析服务端点
func ResolveServiceEndpoints(cfg config.ConsulConf, serviceName string) ([]string, error) {
	if cfg.Address == "" || serviceName == "" {
		return nil, fmt.Errorf("consul address or service name is empty")
	}

	client, err := api.NewClient(&api.Config{
		Address: cfg.Address,
		Token:   cfg.Token,
	})
	if err != nil {
		return nil, fmt.Errorf("create consul client: %w", err)
	}

	services, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("query consul service: %w", err)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("service %s not found in consul", serviceName)
	}

	var endpoints []string
	for _, service := range services {
		address := service.Service.Address
		port := service.Service.Port
		if address == "" {
			address = "127.0.0.1"
		}
		endpoints = append(endpoints, fmt.Sprintf("%s:%d", address, port))
	}

	return endpoints, nil
}
