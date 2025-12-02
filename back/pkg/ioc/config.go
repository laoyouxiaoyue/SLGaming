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
	GetCheckInterval() string
	GetCheckTimeout() string
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
