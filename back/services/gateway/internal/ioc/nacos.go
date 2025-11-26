package ioc

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"SLGaming/back/services/gateway/internal/config"

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

	cacheDir = "tmp/nacos/cache"
	logDir   = "tmp/nacos/log"
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

	client, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  &clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		return nil, fmt.Errorf("new nacos client: %w", err)
	}

	return client, nil
}

// FetchConfig 从 Nacos 拉取配置
func FetchConfig(cli config_client.IConfigClient, cfg config.NacosConf) (string, error) {
	if cli == nil {
		return "", fmt.Errorf("nil nacos client")
	}
	content, err := cli.GetConfig(vo.ConfigParam{
		DataId: cfg.DataId,
		Group:  cfg.Group,
	})
	if err != nil {
		return "", fmt.Errorf("get config: %w", err)
	}
	return content, nil
}

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

func toServerConfig(host string) (constant.ServerConfig, error) {
	scheme := defaultScheme
	contextPath := defaultContextPath
	address := host
	port := defaultPort

	if strings.Contains(host, "://") {
		u, err := url.Parse(host)
		if err != nil {
			return constant.ServerConfig{}, fmt.Errorf("invalid host %s: %w", host, err)
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
		var portStr string
		address, portStr, _ = strings.Cut(address, ":")
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return constant.ServerConfig{}, fmt.Errorf("invalid nacos port %s: %w", portStr, err)
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

// ListenConfig 监听 Nacos 配置变化
func ListenConfig(cli config_client.IConfigClient, cfg config.NacosConf, onChange func(string)) error {
	if cli == nil {
		return fmt.Errorf("nil nacos client")
	}
	return cli.ListenConfig(vo.ConfigParam{
		DataId: cfg.DataId,
		Group:  cfg.Group,
		OnChange: func(namespace, group, dataId, data string) {
			logx.Infof("nacos config updated, dataId=%s, group=%s", dataId, group)
			if onChange != nil {
				onChange(data)
			}
		},
	})
}
