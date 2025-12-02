package ioc

import (
	"SLGaming/back/pkg/ioc"
	"SLGaming/back/services/order/internal/config"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
)

const (
	cacheDir = "tmp/nacos/cache"
	logDir   = "tmp/nacos/log"
)

// InitNacos 根据配置创建一个 Nacos 配置客户端
func InitNacos(cfg config.NacosConf) (config_client.IConfigClient, error) {
	adapter := &ioc.NacosConfigAdapter{
		Hosts:     cfg.Hosts,
		Namespace: cfg.Namespace,
		Group:     cfg.Group,
		DataId:    cfg.DataId,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}
	return ioc.InitNacos(adapter, cacheDir, logDir)
}

// FetchConfig 从 Nacos 获取配置内容
func FetchConfig(client config_client.IConfigClient, cfg config.NacosConf) (string, error) {
	adapter := &ioc.NacosConfigAdapter{
		Hosts:     cfg.Hosts,
		Namespace: cfg.Namespace,
		Group:     cfg.Group,
		DataId:    cfg.DataId,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}
	return ioc.FetchConfig(client, adapter)
}

// ListenConfig 监听 Nacos 配置变化
func ListenConfig(client config_client.IConfigClient, cfg config.NacosConf, onChange func(string)) error {
	adapter := &ioc.NacosConfigAdapter{
		Hosts:     cfg.Hosts,
		Namespace: cfg.Namespace,
		Group:     cfg.Group,
		DataId:    cfg.DataId,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}
	return ioc.ListenConfig(client, adapter, onChange)
}


