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
	Nacos     NacosConf     `json:",optional"` // Nacos 配置
	Consul    ConsulConf    `json:",optional"` // Consul 配置
	Upstream  UpstreamConf  `json:",optional"` // 上游服务配置
	JWT       JWTConf       `json:",optional"` // JWT 配置
	Redis     RedisConf     `json:",optional"` // Redis 配置
	RateLimit RateLimitConf `json:",optional"` // 限流配置
	Upload    UploadConf    `json:",optional"` // 上传配置
	Alipay    AlipayConf    `json:",optional"` // 支付宝配置
	RocketMQ  RocketMQConf  `json:",optional"` // RocketMQ 配置
}

// JWTConf JWT 配置
type JWTConf struct {
	SecretKey            string        `json:",optional"`         // JWT 密钥
	AccessTokenDuration  time.Duration `json:",default=600s"`     // Access Token 过期时间，默认 10 分钟
	RefreshTokenDuration time.Duration `json:",default=1209600s"` // Refresh Token 过期时间，默认 14 天
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
	CodeService  string        `json:",optional"`    // 验证码服务名称（用于 Consul 服务发现）
	UserService  string        `json:",optional"`    // 用户服务名称（用于 Consul 服务发现）
	OrderService string        `json:",optional"`    // 订单服务名称（用于 Consul 服务发现）
	AgentService string        `json:",optional"`    // 智能服务名称（用于 Consul 服务发现）
	RPCTimeout   time.Duration `json:",default=10s"` // RPC 调用超时时间，默认 10 秒
}

// RedisConf Redis 配置
type RedisConf struct {
	redis.RedisConf
}

// RateLimitConf 限流配置
type RateLimitConf struct {
	Enabled   bool                 `json:",default=true"` // 是否启用限流
	GlobalQPS int                  `json:",default=2000"` // 全局 QPS 限制
	Routes    []RouteRateLimitConf `json:",optional"`     // 路由级别的限流配置
}

// UploadConf 上传配置
type UploadConf struct {
	LocalDir         string   `json:",default=uploads"`  // 本地保存目录（相对运行目录）
	BaseURL          string   `json:",default=/uploads"` // 对外访问前缀
	DefaultAvatarURL string   `json:",optional"`         // 审核中默认头像
	MaxSizeMB        int64    `json:",default=5"`        // 单文件最大大小（MB）
	AllowedExt       []string `json:",optional"`         // 允许的扩展名，如 [".jpg",".png"]
}

// AlipayConf 支付宝配置
type AlipayConf struct {
	AppID           string `json:",optional"`      // 应用 AppID
	PrivateKey      string `json:",optional"`      // 应用私钥（PKCS8）
	AlipayPublicKey string `json:",optional"`      // 支付宝公钥
	NotifyURL       string `json:",optional"`      // 异步通知地址
	ReturnURL       string `json:",optional"`      // 同步回跳地址
	IsProduction    bool   `json:",default=false"` // 是否生产环境
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

// RouteRateLimitConf 路由限流配置
type RouteRateLimitConf struct {
	Path             string `json:",optional"` // 路由路径，如 "/api/user/login"
	Method           string `json:",optional"` // HTTP 方法，如 "POST"，为空则匹配所有方法
	GlobalQPS        int    `json:",optional"` // 该路由的全局 QPS 限制
	PerIPQPS         int    `json:",optional"` // 每个 IP 的 QPS 限制
	PerUserQPS       int    `json:",optional"` // 每个用户（已登录）的 QPS 限制
	PerIPPerMinute   int    `json:",optional"` // 每个 IP 每分钟请求数限制
	PerUserPerMinute int    `json:",optional"` // 每个用户每分钟请求数限制
}
