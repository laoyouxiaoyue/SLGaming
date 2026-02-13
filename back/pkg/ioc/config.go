package ioc

import "time"

// ConsulConfig 定义 Consul 配置接口
type ConsulConfig interface {
	GetAddress() string
	GetToken() string
	GetServiceName() string
	GetServiceID() string
	GetServiceAddress() string
	GetServiceTags() []string
	GetServiceMeta() map[string]string
	GetCheckInterval() string
	GetCheckTimeout() string
	GetCheckHTTP() string
}

// NacosConfig 定义 Nacos 配置接口
type NacosConfig interface {
	GetHosts() []string
	GetNamespace() string
	GetGroup() string
	GetDataId() string
	GetUsername() string
	GetPassword() string
}

// RedisConfig 定义 Redis 配置接口
type RedisConfig interface {
	GetHost() string
	GetType() string
	GetPass() string
	GetTls() bool
}

// MySQLConfig 定义 MySQL 配置接口
type MySQLConfig interface {
	GetDSN() string
	GetMaxIdleConns() int
	GetMaxOpenConns() int
	GetConnMaxLifetime() time.Duration
	GetConnMaxIdleTime() time.Duration
}

// RocketMQConfig 定义 RocketMQ 配置接口
type RocketMQConfig interface {
	// NameServer 地址列表，如 []string{"127.0.0.1:9876"}
	GetNameServers() []string
	// 可选：命名空间，用于多租户隔离
	GetNamespace() string
	// 可选：访问凭证（如果开启 ACL）
	GetAccessKey() string
	GetSecretKey() string
}
