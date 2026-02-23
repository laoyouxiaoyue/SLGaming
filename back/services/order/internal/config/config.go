package config

import (
	"time"

	"SLGaming/back/pkg/rpc"

	"github.com/zeromicro/go-zero/zrpc"
)

type NacosConf struct {
	Hosts     []string `json:",optional"`
	Namespace string   `json:",optional"`
	Group     string   `json:",optional"`
	DataId    string   `json:",optional"`
	Username  string   `json:",optional"`
	Password  string   `json:",optional"`
}

type MysqlConf struct {
	DSN             string        `json:",optional"`
	MaxIdleConns    int           `json:",default=10"`
	MaxOpenConns    int           `json:",default=100"`
	ConnMaxLifetime time.Duration `json:",default=300s"`
	ConnMaxIdleTime time.Duration `json:",default=60s"`
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

type UpstreamConf struct {
	UserService string           `json:",optional"`
	RPCTimeout  time.Duration    `json:",default=10s"`
	Retry       rpc.RetryOptions `json:",optional"`
}

type Config struct {
	zrpc.RpcServerConf
	Nacos       NacosConf    `json:",optional"`
	Mysql       MysqlConf    `json:",optional"`
	Consul      ConsulConf   `json:",optional"`
	Upstream    UpstreamConf `json:",optional"`
	RocketMQ    RocketMQConf `json:",optional"`
	MetricsPort int          `json:",optional"`
}

type RocketMQConf struct {
	NameServers []string `json:",optional"`
	Namespace   string   `json:",optional"`
	AccessKey   string   `json:",optional"`
	SecretKey   string   `json:",optional"`
}
