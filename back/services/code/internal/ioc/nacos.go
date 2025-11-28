package ioc

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"SLGaming/back/services/code/internal/config"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultScheme      = "http"
	defaultContextPath = "/nacos"
	defaultPort        = 8848
	defaultTimeoutMs   = 5000

	cacheDir       = "tmp/nacos/cache"
	cacheConfigDir = "tmp/nacos/cache/config"
	logDir         = "tmp/nacos/log"
)

// InitNacos 根据配置创建一个 Nacos 配置客户端。
func InitNacos(cfg config.NacosConf) (config_client.IConfigClient, error) {
	if len(cfg.Hosts) == 0 {
		return nil, fmt.Errorf("nacos hosts is empty")
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:         cfg.Namespace,
		TimeoutMs:           defaultTimeoutMs,
		NotLoadCacheAtStart: true,
		LogDir:              logDir,
		CacheDir:            cacheDir,
		LogLevel:            "info",
		Username:            cfg.Username,
		Password:            cfg.Password,
	}

	serverConfigs, err := buildServerConfigs(cfg.Hosts)
	if err != nil {
		return nil, fmt.Errorf("build server configs: %w", err)
	}

	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("new nacos client: %w", err)
	}

	return configClient, nil
}

// FetchConfig 从 Nacos 获取配置内容
func FetchConfig(client config_client.IConfigClient, cfg config.NacosConf) (string, error) {
	if client == nil {
		return "", fmt.Errorf("nil nacos client")
	}
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: cfg.DataId,
		Group:  cfg.Group,
	})
	if err != nil {
		return "", fmt.Errorf("get config from nacos: %w", err)
	}
	return content, nil
}

// ListenConfig 监听 Nacos 配置变化
func ListenConfig(client config_client.IConfigClient, cfg config.NacosConf, onChange func(string)) error {
	if client == nil {
		return fmt.Errorf("nil nacos client")
	}
	return client.ListenConfig(vo.ConfigParam{
		DataId: cfg.DataId,
		Group:  cfg.Group,
		OnChange: func(namespace, group, dataId, data string) {
			logx.Infof("nacos config updated, dataId=%s, group=%s", dataId, group)
			defer func() {
				if r := recover(); r != nil {
					logx.Errorf("panic recovered in nacos listener: %v", r)
				}
			}()
			if onChange != nil {
				onChange(data)
			}
		},
	})
}

// buildServerConfigs 解析 hosts 列表生成 ServerConfig 切片。
func buildServerConfigs(hosts []string) ([]constant.ServerConfig, error) {
	var serverConfigs []constant.ServerConfig
	for _, h := range hosts {
		if strings.TrimSpace(h) == "" {
			continue
		}
		cfg, err := toServerConfig(strings.TrimSpace(h))
		if err != nil {
			return nil, err
		}
		serverConfigs = append(serverConfigs, cfg)
	}
	if len(serverConfigs) == 0 {
		return nil, fmt.Errorf("no valid nacos hosts")
	}
	return serverConfigs, nil
}

// toServerConfig 将 host 字符串解析为单个 ServerConfig。
func toServerConfig(host string) (constant.ServerConfig, error) {
	scheme := defaultScheme
	contextPath := defaultContextPath
	port := defaultPort
	address := host

	if strings.Contains(host, "://") {
		u, err := url.Parse(host)
		if err != nil {
			return constant.ServerConfig{}, fmt.Errorf("invalid nacos host %s: %w", host, err)
		}
		address = u.Host
		if u.Scheme != "" {
			scheme = u.Scheme
		}
		if u.Path != "" {
			contextPath = u.Path
		}
	}

	if strings.Contains(address, ":") {
		hostPart, portPart, ok := strings.Cut(address, ":")
		if !ok {
			return constant.ServerConfig{}, fmt.Errorf("invalid nacos address %s", address)
		}
		address = hostPart
		p, err := strconv.Atoi(portPart)
		if err != nil {
			return constant.ServerConfig{}, fmt.Errorf("invalid nacos port %s: %w", portPart, err)
		}
		port = p
	}

	return constant.ServerConfig{
		IpAddr:      address,
		Port:        uint64(port),
		Scheme:      scheme,
		ContextPath: contextPath,
	}, nil
}
