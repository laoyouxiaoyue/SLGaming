package ioc

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"SLGaming/back/services/code/internal/config"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"github.com/zeromicro/go-zero/core/logx"
)

type consulRegistrar struct {
	client    *api.Client
	serviceID string
}

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
			GRPC:     fmt.Sprintf("%s:%d", serviceAddr, port),
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
