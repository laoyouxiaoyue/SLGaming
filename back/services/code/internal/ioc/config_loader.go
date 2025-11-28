package ioc

import (
	"os"
	"strings"

	"SLGaming/back/services/code/internal/config"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"gopkg.in/yaml.v3"
)

// LoadConfig 从本地文件和 Nacos 拉取配置，并加载验证码模板
func LoadConfig(configFile, templatesFile string) config.Config {
	var cfg config.Config
	conf.MustLoad(configFile, &cfg)
	loadTemplatesFromFile(&cfg, templatesFile)

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
			loadTemplatesFromFile(&cfg, templatesFile)
			loadTemplatesFromNacos(client, &cfg)
			logx.Infof("load code config from nacos success")
		}
	}

	loadTemplatesFromNacos(client, &cfg)

	go func() {
		if err := ListenConfig(client, cfg.Nacos, func(content string) {
			if strings.TrimSpace(content) == "" {
				return
			}
			logx.Infof("code config updated in nacos, restart service to take effect")
		}); err != nil {
			logx.Errorf("listen nacos config failed: %v", err)
		}
	}()

	return cfg
}

type templateFile struct {
	Templates map[string]config.Template `json:"templates" yaml:"templates"`
}

func loadTemplatesFromFile(cfg *config.Config, path string) {
	if path == "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		logx.Errorf("read template file failed: %v", err)
		return
	}
	var tf templateFile
	if err := yaml.Unmarshal(data, &tf); err != nil {
		logx.Errorf("unmarshal template file failed: %v", err)
		return
	}
	cfg.Template = tf.Templates
	logx.Infof("load template file success %+v", cfg.Template)
}

func loadTemplatesFromNacos(client config_client.IConfigClient, cfg *config.Config) {
	if client == nil || cfg.TemplateNacos.DataId == "" {
		return
	}
	group := cfg.TemplateNacos.Group
	if group == "" {
		group = cfg.Nacos.Group
	}
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: cfg.TemplateNacos.DataId,
		Group:  group,
	})
	if err != nil {
		logx.Errorf("fetch template from nacos failed: %v", err)
		return
	}
	if strings.TrimSpace(content) == "" {
		return
	}
	var tf templateFile
	if err := yaml.Unmarshal([]byte(content), &tf); err != nil {
		logx.Errorf("unmarshal template nacos content failed: %v", err)
		return
	}
	if len(tf.Templates) > 0 {
		cfg.Template = tf.Templates
	}
}
