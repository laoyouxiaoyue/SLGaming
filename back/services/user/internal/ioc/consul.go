package ioc

import (
	"fmt"
	"strconv"

	"SLGaming/back/pkg/ioc"
	"SLGaming/back/services/user/internal/config"
)

// RegisterConsul 注册服务到 Consul（gRPC 服务）
func RegisterConsul(cfg config.ConsulConf, listenOn string, metricsPort int) (*ioc.ConsulRegistrar, error) {
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

	adapter.Service.Meta = map[string]string{
		"metrics_port": strconv.Itoa(metricsPort),
		"protocol":     "grpc",
	}

	hasPrometheusTag := false
	for _, tag := range adapter.Service.Tags {
		if tag == "prometheus" {
			hasPrometheusTag = true
			break
		}
	}
	if !hasPrometheusTag {
		adapter.Service.Tags = append(adapter.Service.Tags, "prometheus")
	}

	adapter.Service.CheckHTTP = fmt.Sprintf("http://%s:%d/metrics", cfg.Service.Address, metricsPort)

	return ioc.RegisterConsul(adapter, listenOn, "grpc")
}
