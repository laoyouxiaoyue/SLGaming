package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// DistributedLock 分布式锁
type DistributedLock struct {
	redis *redis.Redis
}

// NewDistributedLock 创建分布式锁实例
func NewDistributedLock(r *redis.Redis) *DistributedLock {
	if r == nil {
		return nil
	}
	return &DistributedLock{
		redis: r,
	}
}

// LockOptions 锁选项
type LockOptions struct {
	// 锁的过期时间（秒），默认 30 秒
	TTL int
	// 获取锁的重试间隔（毫秒），默认 100ms
	RetryInterval time.Duration
	// 获取锁的最大等待时间（秒），默认 10 秒，0 表示不等待
	MaxWaitTime time.Duration
}

// DefaultLockOptions 返回默认的锁选项
func DefaultLockOptions() *LockOptions {
	return &LockOptions{
		TTL:           30,
		RetryInterval: 100 * time.Millisecond,
		MaxWaitTime:   10 * time.Second,
	}
}

// TryLock 尝试获取锁（非阻塞）
// key: 锁的键
// value: 锁的值（用于标识锁的持有者，建议使用 UUID）
// options: 锁选项，如果为 nil 则使用默认选项
// 返回：是否成功获取锁，错误信息
func (dl *DistributedLock) TryLock(ctx context.Context, key, value string, options *LockOptions) (bool, error) {
	if dl == nil || dl.redis == nil {
		return false, fmt.Errorf("distributed lock not initialized")
	}

	if options == nil {
		options = DefaultLockOptions()
	}

	// 使用 SET NX EX 命令实现分布式锁
	// SET key value NX EX ttl
	// NX: 只在键不存在时设置
	// EX: 设置过期时间（秒）
	lockKey := dl.getLockKey(key)
	ok, err := dl.redis.SetnxEx(lockKey, value, options.TTL)
	if err != nil {
		logx.Errorf("try lock failed: key=%s, error=%v", lockKey, err)
		return false, fmt.Errorf("try lock failed: %w", err)
	}

	if ok {
		logx.Infof("lock acquired: key=%s, value=%s, ttl=%d", lockKey, value, options.TTL)
	}
	return ok, nil
}

// Lock 获取锁（阻塞，会重试）
// key: 锁的键
// value: 锁的值（用于标识锁的持有者，建议使用 UUID）
// options: 锁选项，如果为 nil 则使用默认选项
// 返回：是否成功获取锁，错误信息
func (dl *DistributedLock) Lock(ctx context.Context, key, value string, options *LockOptions) (bool, error) {
	if dl == nil || dl.redis == nil {
		return false, fmt.Errorf("distributed lock not initialized")
	}

	if options == nil {
		options = DefaultLockOptions()
	}

	lockKey := dl.getLockKey(key)
	startTime := time.Now()

	for {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		// 尝试获取锁
		ok, err := dl.redis.SetnxEx(lockKey, value, options.TTL)
		if err != nil {
			logx.Errorf("lock failed: key=%s, error=%v", lockKey, err)
			return false, fmt.Errorf("lock failed: %w", err)
		}

		if ok {
			logx.Infof("lock acquired: key=%s, value=%s, ttl=%d, wait_time=%v",
				lockKey, value, options.TTL, time.Since(startTime))
			return true, nil
		}

		// 检查是否超过最大等待时间
		if options.MaxWaitTime > 0 && time.Since(startTime) >= options.MaxWaitTime {
			logx.Infof("lock timeout: key=%s, max_wait_time=%v", lockKey, options.MaxWaitTime)
			return false, fmt.Errorf("lock timeout after %v", options.MaxWaitTime)
		}

		// 等待后重试
		time.Sleep(options.RetryInterval)
	}
}

// Unlock 释放锁
// key: 锁的键
// value: 锁的值（必须与获取锁时使用的值相同）
// 返回：是否成功释放锁，错误信息
func (dl *DistributedLock) Unlock(ctx context.Context, key, value string) error {
	if dl == nil || dl.redis == nil {
		return fmt.Errorf("distributed lock not initialized")
	}

	lockKey := dl.getLockKey(key)

	// 使用 Lua 脚本确保原子性：只有锁的值匹配时才删除
	// 这样可以防止误删其他进程的锁
	luaScript := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := dl.redis.Eval(luaScript, []string{lockKey}, []string{value})
	if err != nil {
		logx.Errorf("unlock failed: key=%s, value=%s, error=%v", lockKey, value, err)
		return fmt.Errorf("unlock failed: %w", err)
	}

	// result 是删除的键数量，1 表示成功，0 表示锁不存在或值不匹配
	if result == 1 {
		logx.Infof("lock released: key=%s, value=%s", lockKey, value)
	} else {
		// logx.Warnf("unlock failed: key=%s, value=%s, lock not found or value mismatch", lockKey, value)
		return fmt.Errorf("unlock failed: lock not found or value mismatch")
	}

	return nil
}

// Renew 续期锁（延长锁的过期时间）
// key: 锁的键
// value: 锁的值
// ttl: 新的过期时间（秒）
// 返回：是否成功续期，错误信息
func (dl *DistributedLock) Renew(ctx context.Context, key, value string, ttl int) (bool, error) {
	if dl == nil || dl.redis == nil {
		return false, fmt.Errorf("distributed lock not initialized")
	}

	lockKey := dl.getLockKey(key)

	// 使用 Lua 脚本确保原子性：只有锁的值匹配时才续期
	luaScript := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := dl.redis.Eval(luaScript, []string{lockKey}, []string{value, fmt.Sprintf("%d", ttl)})
	if err != nil {
		logx.Errorf("renew lock failed: key=%s, value=%s, error=%v", lockKey, value, err)
		return false, fmt.Errorf("renew lock failed: %w", err)
	}

	if result == 1 {
		logx.Infof("lock renewed: key=%s, value=%s, ttl=%d", lockKey, value, ttl)
		return true, nil
	}

	return false, fmt.Errorf("renew lock failed: lock not found or value mismatch")
}

// getLockKey 获取锁的完整键名
func (dl *DistributedLock) getLockKey(key string) string {
	return fmt.Sprintf("lock:%s", key)
}

// WithLock 使用锁执行函数（自动获取和释放锁）
// 这是一个便捷方法，确保锁在使用后自动释放
func (dl *DistributedLock) WithLock(ctx context.Context, key, value string, options *LockOptions, fn func() error) error {
	if dl == nil || dl.redis == nil {
		return fmt.Errorf("distributed lock not initialized")
	}

	// 获取锁
	acquired, err := dl.Lock(ctx, key, value, options)
	if err != nil {
		return fmt.Errorf("acquire lock failed: %w", err)
	}
	if !acquired {
		return fmt.Errorf("failed to acquire lock: %s", key)
	}

	// 确保释放锁
	defer func() {
		if unlockErr := dl.Unlock(ctx, key, value); unlockErr != nil {
			logx.Errorf("unlock failed in defer: key=%s, error=%v", key, unlockErr)
		}
	}()

	// 执行函数
	return fn()
}
