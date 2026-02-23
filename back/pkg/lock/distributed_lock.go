package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	zeroredis "github.com/zeromicro/go-zero/core/stores/redis"
)

// RedisConfig Redis 配置接口
// 用于从配置中获取 Redis 连接参数
type RedisConfig interface {
	GetHost() string // Redis 地址，如 "127.0.0.1:6379"
	GetType() string // Redis 类型：node 或 cluster
	GetPass() string // Redis 密码
	GetTls() bool    // 是否启用 TLS
}

// DistributedLock 分布式锁管理器
// 基于 redsync 实现，支持单节点 Redis
type DistributedLock struct {
	rs *redsync.Redsync
}

// NewDistributedLock 创建分布式锁实例
//
// 参数：
//   - r: RedisConfig 配置接口，或 *zeroredis.Redis 客户端（不推荐）
//
// 返回：
//   - 成功：*DistributedLock
//   - 失败：nil（错误已记录日志）
//
// 注意：推荐使用 NewDistributedLockFromConfig 创建实例
func NewDistributedLock(r interface{}) *DistributedLock {
	var host, pass string

	// 从 RedisConfig 接口获取配置
	if cfg, ok := r.(RedisConfig); ok {
		host = cfg.GetHost()
		pass = cfg.GetPass()
	} else if _, ok := r.(*zeroredis.Redis); ok {
		// go-zero Redis 客户端无法直接获取配置，不推荐使用
		logx.Errorf("using go-zero redis client, but cannot extract config directly, please use RedisConfig instead")
		return nil
	} else {
		logx.Errorf("invalid redis config type, expected RedisConfig or *redis.Redis")
		return nil
	}

	if host == "" {
		logx.Errorf("redis host is empty, cannot create distributed lock")
		return nil
	}

	// 创建 go-redis 客户端
	opts := &redis.Options{
		Addr: host,
	}
	if pass != "" {
		opts.Password = pass
	}

	client := redis.NewClient(opts)
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	return &DistributedLock{rs: rs}
}

// NewDistributedLockFromConfig 从 RedisConfig 创建分布式锁（推荐方式）
func NewDistributedLockFromConfig(cfg RedisConfig) *DistributedLock {
	if cfg == nil {
		return nil
	}
	return NewDistributedLock(cfg)
}

// LockOptions 锁选项配置
type LockOptions struct {
	// TTL 锁的过期时间（Time To Live）
	// 防止客户端崩溃后锁无法释放，默认 30 秒
	TTL time.Duration

	// RetryInterval 获取锁失败后的重试间隔
	// 默认 100ms
	RetryInterval time.Duration

	// MaxWaitTime 获取锁的最大等待时间
	// 超过此时间将返回超时错误，默认 10 秒
	// 设置为 0 表示无限等待（不推荐）
	MaxWaitTime time.Duration
}

// DefaultLockOptions 返回默认的锁选项
// TTL: 30s, RetryInterval: 100ms, MaxWaitTime: 10s
func DefaultLockOptions() *LockOptions {
	return &LockOptions{
		TTL:           30 * time.Second,
		RetryInterval: 100 * time.Millisecond,
		MaxWaitTime:   10 * time.Second,
	}
}

// LockHandle 锁句柄
//
// 重要：必须使用同一个 LockHandle 实例来释放锁！
//
// 原理说明：
//   - redsync 内部为每个 mutex 生成唯一的 value（UUID）
//   - Redis 释放锁时会校验 value，只有相同 value 才能释放
//   - 因此每次 NewMutex 都会生成新的 value，无法释放其他 mutex 获取的锁
//   - LockHandle 保存了获取锁时的 mutex 实例，确保 Unlock 时使用同一个实例
//
// Redis 锁原理：
//
//	SET lock:key <value> NX PX 30000
//	- NX: 仅当 key 不存在时设置
//	- PX: 设置过期时间（毫秒）
type LockHandle struct {
	mutex  *redsync.Mutex // redsync mutex 实例（包含唯一的 value）
	key    string         // 原始锁键名（不含前缀）
	value  string         // 锁的唯一标识（UUID，用于释放锁时校验）
	expiry time.Time      // 锁的过期时间
	dl     *DistributedLock
}

// Key 返回原始锁键名
func (h *LockHandle) Key() string {
	return h.key
}

// Value 返回锁的唯一标识（UUID）
func (h *LockHandle) Value() string {
	return h.value
}

// Expiry 返回锁的过期时间
func (h *LockHandle) Expiry() time.Time {
	return h.expiry
}

// IsValid 检查锁是否仍然有效（未过期）
func (h *LockHandle) IsValid() bool {
	return h != nil && h.mutex != nil && time.Now().Before(h.expiry)
}

// Unlock 释放锁
//
// 底层 Redis 命令：
//
//	EVAL <script> 1 lock:<key> <value>
//	脚本内容：仅当 Redis 中存储的 value 与参数相同时才删除 key
//
// 参数：
//   - ctx: 上下文
//
// 返回：
//   - nil: 成功释放
//   - error: 释放失败（锁已过期、网络错误等）
//
// 注意：
//   - 锁已过期时会返回 false，但不一定是错误（可能是业务执行时间过长）
//   - 建议使用 defer 确保锁被释放
func (h *LockHandle) Unlock(ctx context.Context) error {
	if h == nil || h.mutex == nil {
		return fmt.Errorf("lock handle is nil")
	}

	ok, err := h.mutex.UnlockContext(ctx)
	if err != nil {
		logx.Errorf("unlock failed: key=%s, error=%v", h.key, err)
		return fmt.Errorf("unlock failed: %w", err)
	}

	if !ok {
		logx.Infof("unlock returned false (lock may have expired): key=%s", h.key)
	}

	logx.Infof("lock released: key=%s", h.key)
	return nil
}

// Renew 续期锁（延长锁的过期时间）
//
// 适用场景：
//   - 任务执行时间超过预期的 TTL
//   - 需要在持有锁期间刷新锁的有效期
//
// 底层 Redis 命令：
//
//	PEXPIRE lock:<key> <ttl_ms>
//	仅当 key 存在且 value 匹配时才续期
//
// 参数：
//   - ctx: 上下文
//   - ttl: 新的过期时间
//
// 返回：
//   - nil: 续期成功
//   - error: 续期失败（锁已过期、已被其他客户端获取等）
//
// 注意：
//   - 续期失败意味着锁已丢失，应停止当前操作
//   - 建议在后台 goroutine 中定期续期
func (h *LockHandle) Renew(ctx context.Context, ttl time.Duration) error {
	if h == nil || h.mutex == nil {
		return fmt.Errorf("lock handle is nil")
	}

	// redsync 的 ExtendContext 方法实现续期
	ok, err := h.mutex.ExtendContext(ctx)
	if err != nil {
		logx.Errorf("renew lock failed: key=%s, error=%v", h.key, err)
		return fmt.Errorf("renew lock failed: %w", err)
	}

	if !ok {
		return fmt.Errorf("renew lock failed: lock may have expired")
	}

	h.expiry = time.Now().Add(ttl)
	logx.Infof("lock renewed: key=%s, new_expiry=%v", h.key, h.expiry)
	return nil
}

// TryLock 尝试获取锁（非阻塞）
//
// 与 Lock 的区别：
//   - TryLock 只尝试一次，不重试
//   - Lock 会按配置重试直到超时
//
// 参数：
//   - ctx: 上下文
//   - key: 锁的键名（会自动添加 "lock:" 前缀）
//   - opts: 锁选项，nil 则使用默认值
//
// 返回：
//   - (*LockHandle, nil): 成功获取锁
//   - (nil, nil): 锁被其他客户端持有（非阻塞返回）
//   - (nil, error): 获取锁失败（网络错误等）
//
// 使用示例：
//
//	handle, err := lock.TryLock(ctx, "order:123", nil)
//	if err != nil { return err }
//	if handle == nil {
//	    // 锁被占用，稍后重试
//	    return errors.New("resource busy")
//	}
//	defer handle.Unlock(ctx)
func (dl *DistributedLock) TryLock(ctx context.Context, key string, opts *LockOptions) (*LockHandle, error) {
	if dl == nil || dl.rs == nil {
		return nil, fmt.Errorf("distributed lock not initialized")
	}

	if opts == nil {
		opts = DefaultLockOptions()
	}

	lockKey := dl.getLockKey(key)
	lockValue := uuid.New().String() // 生成唯一标识

	// 创建 mutex，设置只尝试一次（WithTries(1)）
	mutex := dl.rs.NewMutex(
		lockKey,
		redsync.WithExpiry(opts.TTL),
		redsync.WithTries(1), // 非阻塞：只尝试一次
		redsync.WithRetryDelay(opts.RetryInterval),
		redsync.WithValue(lockValue), // 设置唯一标识
	)

	err := mutex.LockContext(ctx)
	if err != nil {
		if err == redsync.ErrFailed {
			// 锁被占用，返回 nil 表示非阻塞获取失败
			return nil, nil
		}
		logx.Errorf("try lock failed: key=%s, error=%v", lockKey, err)
		return nil, fmt.Errorf("try lock failed: %w", err)
	}

	handle := &LockHandle{
		mutex:  mutex,
		key:    key,
		value:  lockValue,
		expiry: time.Now().Add(opts.TTL),
		dl:     dl,
	}

	logx.Infof("lock acquired: key=%s, value=%s, ttl=%v", lockKey, lockValue, opts.TTL)
	return handle, nil
}

// Lock 获取锁（阻塞，会重试直到超时）
//
// 参数：
//   - ctx: 上下文
//   - key: 锁的键名（会自动添加 "lock:" 前缀）
//   - opts: 锁选项，nil 则使用默认值
//
// 返回：
//   - (*LockHandle, nil): 成功获取锁
//   - (nil, error): 获取锁失败（超时、网络错误等）
//
// 使用示例：
//
//	handle, err := lock.Lock(ctx, "order:123", &lock.LockOptions{
//	    TTL:         30 * time.Second,
//	    MaxWaitTime: 5 * time.Second, // 最多等待 5 秒
//	})
//	if err != nil { return err }
//	defer handle.Unlock(ctx)
func (dl *DistributedLock) Lock(ctx context.Context, key string, opts *LockOptions) (*LockHandle, error) {
	if dl == nil || dl.rs == nil {
		return nil, fmt.Errorf("distributed lock not initialized")
	}

	if opts == nil {
		opts = DefaultLockOptions()
	}

	lockKey := dl.getLockKey(key)
	lockValue := uuid.New().String()

	// 计算重试次数 = MaxWaitTime / RetryInterval
	var tries int
	if opts.MaxWaitTime > 0 {
		tries = int(opts.MaxWaitTime / opts.RetryInterval)
		if tries < 1 {
			tries = 1
		}
	} else {
		tries = 100 // 默认重试 100 次
	}

	// 创建 mutex，设置重试次数
	mutex := dl.rs.NewMutex(
		lockKey,
		redsync.WithExpiry(opts.TTL),
		redsync.WithTries(tries), // 阻塞：按配置重试
		redsync.WithRetryDelay(opts.RetryInterval),
		redsync.WithValue(lockValue),
	)

	// 创建带超时的 context
	lockCtx := ctx
	if opts.MaxWaitTime > 0 {
		var cancel context.CancelFunc
		lockCtx, cancel = context.WithTimeout(ctx, opts.MaxWaitTime)
		defer cancel()
	}

	startTime := time.Now()
	err := mutex.LockContext(lockCtx)
	if err != nil {
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil, fmt.Errorf("lock timeout after %v", opts.MaxWaitTime)
		}
		if err == redsync.ErrFailed {
			return nil, fmt.Errorf("lock acquisition failed after %d tries", tries)
		}
		logx.Errorf("lock failed: key=%s, error=%v", lockKey, err)
		return nil, fmt.Errorf("lock failed: %w", err)
	}

	handle := &LockHandle{
		mutex:  mutex,
		key:    key,
		value:  lockValue,
		expiry: time.Now().Add(opts.TTL),
		dl:     dl,
	}

	logx.Infof("lock acquired: key=%s, value=%s, ttl=%v, wait_time=%v",
		lockKey, lockValue, opts.TTL, time.Since(startTime))
	return handle, nil
}

// WithLock 获取锁并执行函数，自动释放锁（推荐使用）
//
// 优点：
//   - 自动管理锁的生命周期，不会忘记释放
//   - 即使函数 panic，也会通过 defer 释放锁
//
// 参数：
//   - ctx: 上下文
//   - key: 锁的键名
//   - opts: 锁选项，nil 则使用默认值
//   - fn: 需要在锁保护下执行的函数
//
// 返回：
//   - nil: 执行成功
//   - error: 获取锁失败或函数执行失败
//
// 使用示例：
//
//	err := lock.WithLock(ctx, "order:123", nil, func() error {
//	    // 在锁保护下执行的业务逻辑
//	    return doSomething()
//	})
func (dl *DistributedLock) WithLock(ctx context.Context, key string, opts *LockOptions, fn func() error) error {
	// 获取锁
	handle, err := dl.Lock(ctx, key, opts)
	if err != nil {
		return err
	}
	if handle == nil {
		return fmt.Errorf("failed to acquire lock: key=%s", key)
	}
	// 确保释放锁
	defer handle.Unlock(ctx)

	// 执行业务函数
	return fn()
}

// getLockKey 获取锁的完整 Redis 键名
// 自动添加 "lock:" 前缀，避免与其他 Redis key 冲突
func (dl *DistributedLock) getLockKey(key string) string {
	return fmt.Sprintf("lock:%s", key)
}
