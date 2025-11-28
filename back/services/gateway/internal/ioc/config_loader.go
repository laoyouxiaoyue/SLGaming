package ioc

import (
	"strings"

	"SLGaming/back/services/gateway/internal/config"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"gopkg.in/yaml.v3"
)

// LoadConfig 先加载本地文件，再尝试使用 Nacos 覆盖
func LoadConfig(configFile string) config.Config {
	var cfg config.Config
	conf.MustLoad(configFile, &cfg)

	if len(cfg.Nacos.Hosts) == 0 || cfg.Nacos.DataId == "" {
		return cfg
	}

	client, err := InitNacos(cfg.Nacos)
	if err != nil {
		logx.Errorf("init nacos failed, use local config: %v", err)
		return cfg
	}

	cfg = loadFromNacos(client, cfg)
	listenNacos(client, cfg)

	return cfg
}

func loadFromNacos(client config_client.IConfigClient, cfg config.Config) config.Config {
	content, err := FetchConfig(client, cfg.Nacos)
	if err != nil {
		logx.Errorf("fetch nacos config failed, use local config: %v", err)
		return cfg
	}

	if strings.TrimSpace(content) == "" {
		return cfg
	}

	tmp := cfg
	if err := yaml.Unmarshal([]byte(content), &tmp); err != nil {
		logx.Errorf("unmarshal nacos config failed, use local config: %v", err)
		return cfg
	}

	logx.Infof("load gateway config from nacos success")
	return tmp
}

func listenNacos(client config_client.IConfigClient, cfg config.Config) {
	go func() {
		if err := ListenConfig(client, cfg.Nacos, func(content string) {
			if strings.TrimSpace(content) == "" {
				return
			}
			logx.Infof("gateway config updated in nacos, restart service to take effect")
		}); err != nil {
			logx.Errorf("listen nacos config failed: %v", err)
		}
	}()
}
