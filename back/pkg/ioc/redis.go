package ioc

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// InitRedis 根据配置初始化 Redis 连接
// 如果初始化失败会返回错误（不会 panic）
func InitRedis(cfg RedisConfig) (*redis.Redis, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis config is nil")
	}

	host := cfg.GetHost()
	if host == "" {
		return nil, fmt.Errorf("redis host is empty")
	}

	redisConf := redis.RedisConf{
		Host: host,
		Type: cfg.GetType(),
		Pass: cfg.GetPass(),
		Tls:  cfg.GetTls(),
	}

	// 使用 recover 捕获 MustNewRedis 可能的 panic，转换为 error
	var client *redis.Redis
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("init redis panic: %v", r)
			}
		}()
		client = redis.MustNewRedis(redisConf)
	}()

	if err != nil {
		return nil, err
	}

	return client, nil
}
