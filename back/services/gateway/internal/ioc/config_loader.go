package ioc

import (
	"strings"

	"SLGaming/back/services/gateway/internal/config"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"gopkg.in/yaml.v3"
)

// LoadConfig 从本地文件和 Nacos 加载配置
func LoadConfig(configFile string) config.Config {
	var cfg config.Config
	conf.MustLoad(configFile, &cfg)

	// 如果未配置 Nacos，直接返回本地配置
	if len(cfg.Nacos.Hosts) == 0 || cfg.Nacos.DataId == "" {
		return cfg
	}

	// 初始化 Nacos 客户端
	client, err := InitNacos(cfg.Nacos)
	if err != nil {
		logx.Errorf("init nacos failed, use local config: %v", err)
		return cfg
	}

	// 从 Nacos 获取配置
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

	// 监听 Nacos 配置变化
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
