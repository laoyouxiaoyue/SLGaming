package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Nacos         NacosConf           `json:",optional"` // Nacos 配置
	Consul        ConsulConf          `json:",optional"`
	TemplateNacos TemplateNacosConf   `json:",optional"`
	Template      map[string]Template `json:",optional"` // 加载后的模板
	RateLimit     RateLimitConf       `json:",optional"` // 限流配置
}

type Template struct {
	ID               string `json:",optional"`
	CodeLength       int    `json:",default=6"`
	ExpireSeconds    int64  `json:",default=300"`
	MaxDailySends    int    `json:",default=10"` // 每个手机号每日最大发送次数
	ProviderTemplate string `json:",optional"`
	ContentTemplate  string `json:",optional"`
}

// RateLimitConf 限流配置
type RateLimitConf struct {
	IPSendInterval        int `json:",default=60"`  // IP发送间隔（秒），默认60秒
	IPDailyLimit          int `json:",default=100"` // IP每日最大发送次数，默认100次
	PhoneSendInterval     int `json:",default=60"`  // 手机号发送间隔（秒），默认60秒
	VerifyPhoneDailyLimit int `json:",default=50"`  // 验证操作单个手机号每日次数上限
	VerifyIPDailyLimit    int `json:",default=200"` // 验证操作单个IP每日次数上限
}

// NacosConf Nacos 配置结构
type NacosConf struct {
	Hosts     []string `json:",optional"` // Nacos 服务器地址列表
	Namespace string   `json:",optional"` // 命名空间
	Group     string   `json:",optional"` // 配置组
	DataId    string   `json:",optional"` // 配置 DataId
	Username  string   `json:",optional"` // 用户名
	Password  string   `json:",optional"` // 密码
}

type TemplateNacosConf struct {
	DataId string `json:",optional"`
	Group  string `json:",optional"`
}

type ConsulConf struct {
	Address string            `json:",optional"`
	Token   string            `json:",optional"`
	Service ConsulServiceConf `json:",optional"`
}

type ConsulServiceConf struct {
	Name          string   `json:",optional"`
	ID            string   `json:",optional"`
	Address       string   `json:",optional"`
	Tags          []string `json:",optional"`
	CheckInterval string   `json:",optional"`
	CheckTimeout  string   `json:",optional"`
}
