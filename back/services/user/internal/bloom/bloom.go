package bloom

import (
	"context"
	"fmt"
	"strconv"

	"SLGaming/back/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	// UserIDBloomKey 存储用户ID的布隆过滤器键
	UserIDBloomKey = "user:id:bloom"

	// UserCachePrefix 用户信息缓存前缀
	UserCachePrefix = "user:info:"

	// EmptyCacheTTL 空值缓存过期时间（秒）
	EmptyCacheTTL = 300

	// UserCacheTTL 用户信息缓存过期时间（秒）
	UserCacheTTL = 1800
)

// BloomFilter 布隆过滤器服务
type BloomFilter struct {
	redis *redis.Redis
	ctx   context.Context
}

// NewBloomFilter 创建布隆过滤器实例
func NewBloomFilter(svcCtx *svc.ServiceContext) *BloomFilter {
	return &BloomFilter{
		redis: svcCtx.Redis,
		ctx:   context.Background(),
	}
}

// InitUserBloomFilter 初始化用户ID布隆过滤器
func (bf *BloomFilter) InitUserBloomFilter(userIDs []int64) error {
	if bf.redis == nil {
		return fmt.Errorf("redis client not initialized")
	}

	// 批量添加用户ID到布隆过滤器
	for _, id := range userIDs {
		err := bf.AddUserID(id)
		if err != nil {
			return fmt.Errorf("add user id to bloom filter failed: %w", err)
		}
	}
	return nil
}

// AddUserID 添加用户ID到布隆过滤器
func (bf *BloomFilter) AddUserID(userID int64) error {
	if bf.redis == nil {
		return fmt.Errorf("redis client not initialized")
	}

	// 使用Redis布隆过滤器命令 BF.ADD
	_, err := bf.redis.Eval(`
		return redis.call('BF.ADD', KEYS[1], ARGV[1])
	`, []string{UserIDBloomKey}, strconv.FormatInt(userID, 10))
	return err
}

// MightContainUserID 判断用户ID是否可能存在
func (bf *BloomFilter) MightContainUserID(userID int64) (bool, error) {
	if bf.redis == nil {
		// Redis未初始化时，降级为始终返回true，避免影响正常流程
		return true, nil
	}

	// 使用Redis布隆过滤器命令 BF.EXISTS
	result, err := bf.redis.Eval(`
		return redis.call('BF.EXISTS', KEYS[1], ARGV[1])
	`, []string{UserIDBloomKey}, strconv.FormatInt(userID, 10))
	if err != nil {
		// 查询失败时，降级为返回true，避免影响正常流程
		return true, nil
	}

	// redis may return int64 or int depending on driver; handle both
	switch v := result.(type) {
	case int64:
		return v == 1, nil
	case int:
		return v == 1, nil
	default:
		return true, nil
	}
}

// GetUserCacheKey 获取用户缓存键
func GetUserCacheKey(userID int64) string {
	return UserCachePrefix + strconv.FormatInt(userID, 10)
}
