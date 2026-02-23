package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Nacos         NacosConf           `json:",optional"`
	Consul        ConsulConf          `json:",optional"`
	TemplateNacos TemplateNacosConf   `json:",optional"`
	Template      map[string]Template `json:",optional"`
	RateLimit     RateLimitConf       `json:",optional"`
	MetricsPort   int                 `json:",default=9092"`
}

type Template struct {
	ID               string `json:",optional"`
	CodeLength       int    `json:",default=6"`
	ExpireSeconds    int64  `json:",default=300"`
	MaxDailySends    int    `json:",default=10"`
	ProviderTemplate string `json:",optional"`
	ContentTemplate  string `json:",optional"`
}

type RateLimitConf struct {
	IPSendInterval        int `json:",default=60"`
	IPDailyLimit          int `json:",default=100"`
	PhoneSendInterval     int `json:",default=60"`
	VerifyPhoneDailyLimit int `json:",default=50"`
	VerifyIPDailyLimit    int `json:",default=200"`
}

type NacosConf struct {
	Hosts     []string `json:",optional"`
	Namespace string   `json:",optional"`
	Group     string   `json:",optional"`
	DataId    string   `json:",optional"`
	Username  string   `json:",optional"`
	Password  string   `json:",optional"`
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
