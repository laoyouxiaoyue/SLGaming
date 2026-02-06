package config

import (
	"time"

	"github.com/zeromicro/go-zero/zrpc"
)

// NacosConf Nacos 配置结构
type NacosConf struct {
	Hosts     []string `json:",optional"` // Nacos 服务器地址列表
	Namespace string   `json:",optional"` // 命名空间
	Group     string   `json:",optional"` // 配置组
	DataId    string   `json:",optional"` // 配置 DataId
	Username  string   `json:",optional"` // 用户名
	Password  string   `json:",optional"` // 密码
}

type Config struct {
	zrpc.RpcServerConf
	Nacos    NacosConf    `json:",optional"` // Nacos 配置
	Mysql    MysqlConf    `json:",optional"` // Mysql 配置
	Consul   ConsulConf   `json:",optional"`
	Upstream UpstreamConf `json:",optional"` // 上游服务配置

	// RocketMQ 消息队列配置（用于钱包变动、订单结算等异步处理）
	RocketMQ RocketMQConf `json:",optional"`
}

// UpstreamConf 上游服务配置
type UpstreamConf struct {
	AgentService string `json:",optional"` // 智能服务名称（用于 Consul 服务发现）
}

// MysqlConf MySQL 数据库配置
type MysqlConf struct {
	DSN             string        `json:",optional"`     // 数据源，user:pass@tcp(host:port)/db?charset=utf8mb4&parseTime=true&loc=Local
	MaxIdleConns    int           `json:",default=10"`   // 最大空闲连接数
	MaxOpenConns    int           `json:",default=100"`  // 最大打开连接数
	ConnMaxLifetime time.Duration `json:",default=300s"` // 连接最大生命周期
	ConnMaxIdleTime time.Duration `json:",default=60s"`  // 连接最大空闲时间
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

// RocketMQConf RocketMQ 配置结构
type RocketMQConf struct {
	// NameServer 地址列表，如 ["127.0.0.1:9876"]
	NameServers []string `json:",optional"`
	// 可选：命名空间，用于环境/租户隔离
	Namespace string `json:",optional"`
	// 可选：访问凭证（如果开启 ACL）
	AccessKey string `json:",optional"`
	SecretKey string `json:",optional"`
}
