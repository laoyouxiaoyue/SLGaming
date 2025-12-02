package ioc

import (
	"SLGaming/back/pkg/ioc"
	"SLGaming/back/services/user/internal/config"
)

// RegisterConsul 注册服务到 Consul（gRPC 服务）
func RegisterConsul(cfg config.ConsulConf, listenOn string) (*ioc.ConsulRegistrar, error) {
	adapter := &ioc.ConsulConfigAdapter{
		Address: cfg.Address,
		Token:   cfg.Token,
	}
	adapter.Service.Name = cfg.Service.Name
	adapter.Service.ID = cfg.Service.ID
	adapter.Service.Address = cfg.Service.Address
	adapter.Service.Tags = cfg.Service.Tags
	adapter.Service.CheckInterval = cfg.Service.CheckInterval
	adapter.Service.CheckTimeout = cfg.Service.CheckTimeout

	return ioc.RegisterConsul(adapter, listenOn, "grpc")
}
