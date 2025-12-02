// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Nacos    NacosConf    `json:",optional"` // Nacos 配置
	Consul   ConsulConf   `json:",optional"` // Consul 配置
	Upstream UpstreamConf `json:",optional"` // 上游服务配置
	JWT      JWTConf      `json:",optional"` // JWT 配置
	Redis    RedisConf    `json:",optional"` // Redis 配置
}

// JWTConf JWT 配置
type JWTConf struct {
	SecretKey            string        `json:",optional"`        // JWT 密钥
	AccessTokenDuration  time.Duration `json:",default=900s"`    // Access Token 过期时间，默认 15 分钟
	RefreshTokenDuration time.Duration `json:",default=604800s"` // Refresh Token 过期时间，默认 7 天
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

// ConsulConf Consul 配置结构
type ConsulConf struct {
	Address string            `json:",optional"`
	Token   string            `json:",optional"`
	Service ConsulServiceConf `json:",optional"`
}

// ConsulServiceConf Consul 服务配置
type ConsulServiceConf struct {
	Name          string   `json:",optional"`
	ID            string   `json:",optional"`
	Address       string   `json:",optional"`
	Tags          []string `json:",optional"`
	CheckInterval string   `json:",optional"`
	CheckTimeout  string   `json:",optional"`
}

// UpstreamConf 上游服务配置
type UpstreamConf struct {
	CodeService  string `json:",optional"` // 验证码服务名称（用于 Consul 服务发现）
	UserService  string `json:",optional"` // 用户服务名称（用于 Consul 服务发现）
	OrderService string `json:",optional"` // 订单服务名称（用于 Consul 服务发现）
}

// RedisConf Redis 配置
type RedisConf struct {
	redis.RedisConf
}
