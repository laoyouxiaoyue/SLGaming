package cache

import (
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// Manager 通用缓存管理器
type Manager struct {
	redis *redis.Redis
}

// NewManager 创建缓存管理器实例
func NewManager(redis *redis.Redis) *Manager {
	return &Manager{
		redis: redis,
	}
}

// Get 获取缓存
func (m *Manager) Get(key string) (string, error) {
	if m.redis == nil {
		return "", redis.Nil
	}
	return m.redis.Get(key)
}

// Set 设置缓存
func (m *Manager) Set(key string, value interface{}, expire time.Duration) error {
	if m.redis == nil {
		return nil
	}
	// 将interface{}转换为string
	strValue, ok := value.(string)
	if !ok {
		// 如果不是string类型，尝试使用fmt.Sprintf转换
		strValue = fmt.Sprintf("%v", value)
	}
	return m.redis.Setex(key, strValue, int(expire.Seconds()))
}

// Incr 增加缓存值
func (m *Manager) Incr(key string) (int64, error) {
	if m.redis == nil {
		return 0, nil
	}
	return m.redis.Incr(key)
}

// Decr 减少缓存值
func (m *Manager) Decr(key string) (int64, error) {
	if m.redis == nil {
		return 0, nil
	}
	return m.redis.Decr(key)
}

// Delete 删除缓存
func (m *Manager) Delete(key string) error {
	if m.redis == nil {
		return nil
	}
	_, err := m.redis.Del(key)
	return err
}

// Expire 设置缓存过期时间
func (m *Manager) Expire(key string, expire time.Duration) error {
	if m.redis == nil {
		return nil
	}
	return m.redis.Expire(key, int(expire.Seconds()))
}

// DeleteMultiple 批量删除缓存
func (m *Manager) DeleteMultiple(keys []string) error {
	if m.redis == nil || len(keys) == 0 {
		return nil
	}
	for _, key := range keys {
		_, err := m.redis.Del(key)
		if err != nil {
			return err
		}
	}
	return nil
}
