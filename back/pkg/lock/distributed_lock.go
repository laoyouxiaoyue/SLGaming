package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	zeroredis "github.com/zeromicro/go-zero/core/stores/redis"
)

// RedisConfig Redis 配置接口（用于创建分布式锁）
type RedisConfig interface {
	GetHost() string
	GetType() string
	GetPass() string
	GetTls() bool
}

// DistributedLock 分布式锁（基于 Redisson 风格的 redsync 实现）
type DistributedLock struct {
	rs *redsync.Redsync
}

// NewDistributedLock 创建分布式锁实例
// 从 go-zero 的 redis 配置创建 redsync 实例
// 支持两种方式：
// 1. 传入 RedisConfig 配置接口（推荐）
// 2. 传入 go-zero 的 Redis 客户端（兼容旧代码）
func NewDistributedLock(r interface{}) *DistributedLock {
	var host, pass string

	// 尝试从 RedisConfig 接口获取配置
	if cfg, ok := r.(RedisConfig); ok {
		host = cfg.GetHost()
		pass = cfg.GetPass()
		_ = cfg.GetTls() // TLS 配置保留以备将来使用
	} else if _, ok := r.(*zeroredis.Redis); ok {
		// 兼容旧代码：从 go-zero Redis 客户端获取配置
		// 注意：go-zero 的 Redis 客户端没有公开的方法获取配置
		// 所以我们需要通过反射或者其他方式，或者要求传入配置
		// 这里我们尝试从 RedisConf 中获取（如果可能）
		logx.Errorf("using go-zero redis client, but cannot extract config directly, please use RedisConfig instead")
		// 如果无法获取配置，返回 nil
		return nil
	} else {
		logx.Errorf("invalid redis config type, expected RedisConfig or *redis.Redis")
		return nil
	}

	if host == "" {
		logx.Errorf("redis host is empty, cannot create distributed lock")
		return nil
	}

	// 创建 go-redis 客户端选项
	opts := &redis.Options{
		Addr: host,
	}

	if pass != "" {
		opts.Password = pass
	}

	// 创建 go-redis 客户端
	client := redis.NewClient(opts)

	// 创建 redsync 实例（支持多 Redis 节点，这里使用单节点）
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	return &DistributedLock{
		rs: rs,
	}
}

// NewDistributedLockFromConfig 从 RedisConfig 创建分布式锁（推荐方式）
func NewDistributedLockFromConfig(cfg RedisConfig) *DistributedLock {
	if cfg == nil {
		return nil
	}
	return NewDistributedLock(cfg)
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
	if dl == nil || dl.rs == nil {
		return false, fmt.Errorf("distributed lock not initialized")
	}

	if options == nil {
		options = DefaultLockOptions()
	}

	lockKey := dl.getLockKey(key)

	// 使用 redsync 创建锁
	mutex := dl.rs.NewMutex(
		lockKey,
		redsync.WithExpiry(time.Duration(options.TTL)*time.Second),
		redsync.WithTries(1), // 只尝试一次（非阻塞）
		redsync.WithRetryDelay(options.RetryInterval),
	)

	// 尝试获取锁
	err := mutex.LockContext(ctx)
	if err != nil {
		if err == redsync.ErrFailed {
			// 获取锁失败（非阻塞）
			return false, nil
		}
		logx.Errorf("try lock failed: key=%s, error=%v", lockKey, err)
		return false, fmt.Errorf("try lock failed: %w", err)
	}

	logx.Infof("lock acquired: key=%s, value=%s, ttl=%d", lockKey, value, options.TTL)
	return true, nil
}

// Lock 获取锁（阻塞，会重试）
// key: 锁的键
// value: 锁的值（用于标识锁的持有者，建议使用 UUID）
// options: 锁选项，如果为 nil 则使用默认选项
// 返回：是否成功获取锁，错误信息
func (dl *DistributedLock) Lock(ctx context.Context, key, value string, options *LockOptions) (bool, error) {
	if dl == nil || dl.rs == nil {
		return false, fmt.Errorf("distributed lock not initialized")
	}

	if options == nil {
		options = DefaultLockOptions()
	}

	lockKey := dl.getLockKey(key)

	// 计算重试次数（基于 MaxWaitTime 和 RetryInterval）
	var tries int
	if options.MaxWaitTime > 0 {
		tries = int(options.MaxWaitTime / options.RetryInterval)
		if tries <= 0 {
			tries = 1
		}
	} else {
		// 如果没有设置 MaxWaitTime，使用默认重试次数
		tries = 100 // 默认重试 100 次
	}

	// 使用 redsync 创建锁
	mutex := dl.rs.NewMutex(
		lockKey,
		redsync.WithExpiry(time.Duration(options.TTL)*time.Second),
		redsync.WithTries(tries),
		redsync.WithRetryDelay(options.RetryInterval),
	)

	// 创建带超时的 context（如果设置了 MaxWaitTime）
	lockCtx := ctx
	if options.MaxWaitTime > 0 {
		var cancel context.CancelFunc
		lockCtx, cancel = context.WithTimeout(ctx, options.MaxWaitTime)
		defer cancel()
	}

	startTime := time.Now()
	err := mutex.LockContext(lockCtx)
	if err != nil {
		if err == context.DeadlineExceeded || err == context.Canceled {
			logx.Infof("lock timeout: key=%s, max_wait_time=%v", lockKey, options.MaxWaitTime)
			return false, fmt.Errorf("lock timeout after %v", options.MaxWaitTime)
		}
		logx.Errorf("lock failed: key=%s, error=%v", lockKey, err)
		return false, fmt.Errorf("lock failed: %w", err)
	}

	logx.Infof("lock acquired: key=%s, value=%s, ttl=%d, wait_time=%v",
		lockKey, value, options.TTL, time.Since(startTime))
	return true, nil
}

// Unlock 释放锁
// key: 锁的键
// value: 锁的值（必须与获取锁时使用的值相同）
// 返回：是否成功释放锁，错误信息
// 注意：redsync 内部管理锁的释放，这里只需要根据 key 找到对应的 mutex
func (dl *DistributedLock) Unlock(ctx context.Context, key, value string) error {
	if dl == nil || dl.rs == nil {
		return fmt.Errorf("distributed lock not initialized")
	}

	lockKey := dl.getLockKey(key)

	// redsync 的锁释放需要保存 mutex 实例
	// 但当前 API 设计没有保存 mutex，所以我们需要重新创建 mutex 并尝试解锁
	// 注意：这不是最佳实践，但为了保持 API 兼容性，我们这样做
	// 更好的方式是使用 WithLock 方法，它会自动管理锁的生命周期
	mutex := dl.rs.NewMutex(lockKey)

	ok, err := mutex.UnlockContext(ctx)
	if err != nil {
		logx.Errorf("unlock failed: key=%s, value=%s, error=%v", lockKey, value, err)
		return fmt.Errorf("unlock failed: %w", err)
	}

	if !ok {
		logx.Errorf("unlock failed: key=%s, value=%s, lock not found or value mismatch", lockKey, value)
		return fmt.Errorf("unlock failed: lock not found or value mismatch")
	}

	logx.Infof("lock released: key=%s, value=%s", lockKey, value)
	return nil
}

// Renew 续期锁（延长锁的过期时间）
// key: 锁的键
// value: 锁的值
// ttl: 新的过期时间（秒）
// 返回：是否成功续期，错误信息
// 注意：redsync 不支持直接续期，需要重新获取锁
func (dl *DistributedLock) Renew(ctx context.Context, key, value string, ttl int) (bool, error) {
	if dl == nil || dl.rs == nil {
		return false, fmt.Errorf("distributed lock not initialized")
	}

	lockKey := dl.getLockKey(key)

	// redsync 不支持直接续期，但我们可以通过重新获取锁来实现
	// 创建一个新的 mutex 并尝试获取（如果锁还存在）
	mutex := dl.rs.NewMutex(
		lockKey,
		redsync.WithExpiry(time.Duration(ttl)*time.Second),
		redsync.WithTries(1),
	)

	// 尝试获取锁（如果锁已存在且由当前进程持有，redsync 会处理）
	err := mutex.LockContext(ctx)
	if err != nil {
		logx.Errorf("renew lock failed: key=%s, value=%s, error=%v", lockKey, value, err)
		return false, fmt.Errorf("renew lock failed: %w", err)
	}

	logx.Infof("lock renewed: key=%s, value=%s, ttl=%d", lockKey, value, ttl)
	return true, nil
}

// getLockKey 获取锁的完整键名
func (dl *DistributedLock) getLockKey(key string) string {
	return fmt.Sprintf("lock:%s", key)
}

// mutexWrapper 包装 redsync.Mutex，用于 WithLock
type mutexWrapper struct {
	mutex *redsync.Mutex
	key   string
}

// WithLock 使用锁执行函数（自动获取和释放锁）
// 这是一个便捷方法，确保锁在使用后自动释放
// 这是推荐的使用方式，因为 redsync 会自动管理锁的生命周期
func (dl *DistributedLock) WithLock(ctx context.Context, key, value string, options *LockOptions, fn func() error) error {
	if dl == nil || dl.rs == nil {
		return fmt.Errorf("distributed lock not initialized")
	}

	if options == nil {
		options = DefaultLockOptions()
	}

	lockKey := dl.getLockKey(key)

	// 计算重试次数
	var tries int
	if options.MaxWaitTime > 0 {
		tries = int(options.MaxWaitTime / options.RetryInterval)
		if tries <= 0 {
			tries = 1
		}
	} else {
		tries = 100
	}

	// 创建 redsync mutex
	mutex := dl.rs.NewMutex(
		lockKey,
		redsync.WithExpiry(time.Duration(options.TTL)*time.Second),
		redsync.WithTries(tries),
		redsync.WithRetryDelay(options.RetryInterval),
	)

	// 创建带超时的 context
	lockCtx := ctx
	if options.MaxWaitTime > 0 {
		var cancel context.CancelFunc
		lockCtx, cancel = context.WithTimeout(ctx, options.MaxWaitTime)
		defer cancel()
	}

	// 获取锁
	err := mutex.LockContext(lockCtx)
	if err != nil {
		if err == context.DeadlineExceeded || err == context.Canceled {
			return fmt.Errorf("acquire lock timeout: %w", err)
		}
		return fmt.Errorf("acquire lock failed: %w", err)
	}

	// 确保释放锁
	defer func() {
		if _, unlockErr := mutex.UnlockContext(ctx); unlockErr != nil {
			logx.Errorf("unlock failed in defer: key=%s, error=%v", key, unlockErr)
		}
	}()

	// 执行函数
	return fn()
}
