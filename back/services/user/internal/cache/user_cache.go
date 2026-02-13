package cache

import (
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// UserCache 用户相关缓存服务
type UserCache struct {
	manager *Manager
}

// NewUserCache 创建用户缓存服务实例
func NewUserCache(manager *Manager) *UserCache {
	return &UserCache{
		manager: manager,
	}
}

// GetFollowerCount 获取用户粉丝数
func (c *UserCache) GetFollowerCount(userId int64) (int64, error) {
	key := fmt.Sprintf(FollowerCountKey, userId)
	val, err := c.manager.Get(key)
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}

	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// SetFollowerCount 设置用户粉丝数
func (c *UserCache) SetFollowerCount(userId int64, count int64) error {
	key := fmt.Sprintf(FollowerCountKey, userId)
	return c.manager.Set(key, count, CountCacheExpire)
}

// GetFollowingCount 获取用户关注数
func (c *UserCache) GetFollowingCount(userId int64) (int64, error) {
	key := fmt.Sprintf(FollowingCountKey, userId)
	val, err := c.manager.Get(key)
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}

	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// SetFollowingCount 设置用户关注数
func (c *UserCache) SetFollowingCount(userId int64, count int64) error {
	key := fmt.Sprintf(FollowingCountKey, userId)
	return c.manager.Set(key, count, CountCacheExpire)
}

// IncrFollowerCount 增加用户粉丝数
func (c *UserCache) IncrFollowerCount(userId int64) error {
	key := fmt.Sprintf(FollowerCountKey, userId)
	_, err := c.manager.Incr(key)
	if err != nil {
		return err
	}

	// 重新设置过期时间
	return c.manager.Expire(key, CountCacheExpire)
}

// DecrFollowerCount 减少用户粉丝数（使用Lua脚本确保不小于0）
func (c *UserCache) DecrFollowerCount(userId int64) error {
	key := fmt.Sprintf(FollowerCountKey, userId)

	// Lua脚本：如果当前值存在且大于0，则减1；否则设置为0
	script := `
		local current = redis.call('GET', KEYS[1])
		if current and tonumber(current) > 0 then
			redis.call('DECR', KEYS[1])
			redis.call('EXPIRE', KEYS[1], ARGV[1])
			return tonumber(current) - 1
		else
			redis.call('SET', KEYS[1], 0)
			redis.call('EXPIRE', KEYS[1], ARGV[1])
			return 0
		end
	`
	_, err := c.manager.Eval(script, []string{key}, int(CountCacheExpire.Seconds()))
	return err
}

// IncrFollowingCount 增加用户关注数
func (c *UserCache) IncrFollowingCount(userId int64) error {
	key := fmt.Sprintf(FollowingCountKey, userId)
	_, err := c.manager.Incr(key)
	if err != nil {
		return err
	}

	// 重新设置过期时间
	return c.manager.Expire(key, CountCacheExpire)
}

// DecrFollowingCount 减少用户关注数（使用Lua脚本确保不小于0）
func (c *UserCache) DecrFollowingCount(userId int64) error {
	key := fmt.Sprintf(FollowingCountKey, userId)

	// Lua脚本：如果当前值存在且大于0，则减1；否则设置为0
	script := `
		local current = redis.call('GET', KEYS[1])
		if current and tonumber(current) > 0 then
			redis.call('DECR', KEYS[1])
			redis.call('EXPIRE', KEYS[1], ARGV[1])
			return tonumber(current) - 1
		else
			redis.call('SET', KEYS[1], 0)
			redis.call('EXPIRE', KEYS[1], ARGV[1])
			return 0
		end
	`
	_, err := c.manager.Eval(script, []string{key}, int(CountCacheExpire.Seconds()))
	return err
}

// DeleteCountCache 删除用户计数缓存
func (c *UserCache) DeleteCountCache(userId int64) error {
	followerKey := fmt.Sprintf(FollowerCountKey, userId)
	followingKey := fmt.Sprintf(FollowingCountKey, userId)

	return c.manager.DeleteMultiple([]string{followerKey, followingKey})
}

// GetUserInfo 获取用户信息缓存
func (c *UserCache) GetUserInfo(userId int64) (string, error) {
	key := fmt.Sprintf(UserInfoKey, userId)
	return c.manager.Get(key)
}

// SetUserInfo 设置用户信息缓存
func (c *UserCache) SetUserInfo(userId int64, value interface{}) error {
	key := fmt.Sprintf(UserInfoKey, userId)
	return c.manager.Set(key, value, UserInfoExpire)
}

// DeleteUserInfo 删除用户信息缓存
func (c *UserCache) DeleteUserInfo(userId int64) error {
	key := fmt.Sprintf(UserInfoKey, userId)
	return c.manager.Delete(key)
}
