package ioc

import (
	"strings"

	"SLGaming/back/services/gateway/internal/config"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"gopkg.in/yaml.v3"
)

// LoadConfig 读取本地配置，并尝试通过 Nacos 覆盖。
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

	if content, err := FetchConfig(client, cfg.Nacos); err != nil {
		logx.Errorf("fetch nacos config failed, use local config: %v", err)
	} else if strings.TrimSpace(content) != "" {
		tmp := cfg
		if err := yaml.Unmarshal([]byte(content), &tmp); err != nil {
			logx.Errorf("unmarshal nacos config failed, use local config: %v", err)
		} else {
			cfg = tmp
			logx.Infof("load gateway config from nacos success")
		}
	}

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

	return cfg
}
