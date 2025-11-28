// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"time"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Nacos    NacosConf    `json:",optional"`
	Consul   ConsulConf   `json:",optional"`
	Upstream UpstreamConf `json:",optional"`
	JWT      JWTConf      `json:",optional"`
}

type NacosConf struct {
	Hosts     []string `json:",optional"`
	Namespace string   `json:",optional"`
	Group     string   `json:",optional"`
	DataId    string   `json:",optional"`
	Username  string   `json:",optional"`
	Password  string   `json:",optional"`
}

type ConsulConf struct {
	Address string `json:",optional"`
	Token   string `json:",optional"`
}

type UpstreamConf struct {
	CodeService string `json:",optional"`
	UserService string `json:",optional"`
}

type JWTConf struct {
	Secret string        `json:",default=slgaming-gateway-secret"`
	TTL    time.Duration `json:",default=24h"`
}
