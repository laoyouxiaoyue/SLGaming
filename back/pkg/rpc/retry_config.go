// Package rpc 提供RPC客户端的增强功能，包括：
//   - 自动重试机制（支持指数退避）
//   - 动态服务发现与客户端更新
//   - gRPC拦截器集成
//
// 典型使用场景：
//
//	client, err := rpc.NewDynamicRPCClientOrFallback(consulConfig, rpc.DynamicClientOptions{
//	    ServiceName: "user-rpc",
//	    Timeout:     10 * time.Second,
//	    Retry:       rpc.DefaultRetryOptions(),
//	})
package rpc

import (
	"time"

	"google.golang.org/grpc/codes"
)

// 默认重试配置常量
const (
	// DefaultMaxRetries 最大重试次数，默认3次
	// 意味着总共会执行4次请求（首次 + 3次重试）
	DefaultMaxRetries = 3

	// DefaultInitialBackoff 初始退避时间，默认100毫秒
	// 第一次重试前等待100ms
	DefaultInitialBackoff = 100 * time.Millisecond

	// DefaultMaxBackoff 最大退避时间，默认5秒
	// 退避时间不会超过此值
	DefaultMaxBackoff = 5 * time.Second

	// DefaultBackoffMultiplier 退避时间倍数，默认2.0
	// 每次重试的退避时间 = 上次退避时间 × 倍数
	// 例如: 100ms → 200ms → 400ms → 800ms → ...
	DefaultBackoffMultiplier = 2.0

	// DefaultRetryTimeout 重试总超时时间，默认30秒
	// 包括所有重试的时间总和
	DefaultRetryTimeout = 30 * time.Second
)

// DefaultRetryableCodes 默认可重试的gRPC错误码
//
// 这些错误码代表瞬时故障，通常可以通过重试恢复：
//   - Unavailable: 服务不可用（服务重启、网络抖动）
//   - DeadlineExceeded: 请求超时（服务负载高）
//   - Aborted: 事务中止（并发冲突）
//   - ResourceExhausted: 资源耗尽（限流）
//
// 不在此列表中的错误码（如 InvalidArgument、NotFound）不会被重试
var DefaultRetryableCodes = []codes.Code{
	codes.Unavailable,
	codes.DeadlineExceeded,
	codes.Aborted,
	codes.ResourceExhausted,
}

// RetryOptions 重试配置选项
//
// 配置示例（YAML）:
//
//	Upstream:
//	  Retry:
//	    MaxRetries: 3
//	    InitialBackoff: 100ms
//	    MaxBackoff: 5s
//	    BackoffMultiplier: 2.0
//	    RetryTimeout: 30s
type RetryOptions struct {
	// MaxRetries 最大重试次数
	// 设置为0表示禁用重试
	MaxRetries int `json:",default=3"`

	// InitialBackoff 初始退避时间
	// 第一次重试前的等待时间
	InitialBackoff time.Duration `json:",default=100ms"`

	// MaxBackoff 最大退避时间上限
	// 退避时间不会超过此值，防止等待过长
	MaxBackoff time.Duration `json:",default=5s"`

	// BackoffMultiplier 退避时间倍数
	// 用于计算指数退避：backoff = backoff × multiplier
	BackoffMultiplier float64 `json:",default=2.0"`

	// RetryTimeout 重试总超时时间
	// 包括所有重试尝试的总时间
	RetryTimeout time.Duration `json:",default=30s"`

	// RetryableCodes 可重试的gRPC错误码列表
	// 为空时使用 DefaultRetryableCodes
	RetryableCodes []codes.Code `json:",optional"`
}

// DefaultRetryOptions 返回默认的重试配置
//
// 返回配置：
//   - MaxRetries: 3
//   - InitialBackoff: 100ms
//   - MaxBackoff: 5s
//   - BackoffMultiplier: 2.0
//   - RetryTimeout: 30s
//   - RetryableCodes: DefaultRetryableCodes
func DefaultRetryOptions() RetryOptions {
	return RetryOptions{
		MaxRetries:        DefaultMaxRetries,
		InitialBackoff:    DefaultInitialBackoff,
		MaxBackoff:        DefaultMaxBackoff,
		BackoffMultiplier: DefaultBackoffMultiplier,
		RetryTimeout:      DefaultRetryTimeout,
		RetryableCodes:    DefaultRetryableCodes,
	}
}

// WithMaxRetries 设置最大重试次数（链式配置）
//
// 示例:
//
//	opts := rpc.DefaultRetryOptions().WithMaxRetries(5)
func (o RetryOptions) WithMaxRetries(n int) RetryOptions {
	o.MaxRetries = n
	return o
}

// WithInitialBackoff 设置初始退避时间（链式配置）
func (o RetryOptions) WithInitialBackoff(d time.Duration) RetryOptions {
	o.InitialBackoff = d
	return o
}

// WithMaxBackoff 设置最大退避时间（链式配置）
func (o RetryOptions) WithMaxBackoff(d time.Duration) RetryOptions {
	o.MaxBackoff = d
	return o
}

// WithBackoffMultiplier 设置退避倍数（链式配置）
func (o RetryOptions) WithBackoffMultiplier(m float64) RetryOptions {
	o.BackoffMultiplier = m
	return o
}

// WithRetryTimeout 设置重试总超时时间（链式配置）
func (o RetryOptions) WithRetryTimeout(d time.Duration) RetryOptions {
	o.RetryTimeout = d
	return o
}

// WithRetryableCodes 设置可重试的错误码（链式配置）
//
// 示例 - 只重试 Unavailable 错误:
//
//	opts := rpc.DefaultRetryOptions().WithRetryableCodes(codes.Unavailable)
func (o RetryOptions) WithRetryableCodes(codes ...codes.Code) RetryOptions {
	o.RetryableCodes = codes
	return o
}

// IsRetryable 检查给定的gRPC错误码是否可重试
//
// 如果 RetryableCodes 为空，使用 DefaultRetryableCodes
func (o RetryOptions) IsRetryable(code codes.Code) bool {
	if len(o.RetryableCodes) == 0 {
		o.RetryableCodes = DefaultRetryableCodes
	}
	for _, c := range o.RetryableCodes {
		if c == code {
			return true
		}
	}
	return false
}

// CalculateBackoff 计算指定重试次数的退避时间
//
// 使用指数退避算法：
//
//	第1次重试: InitialBackoff
//	第2次重试: InitialBackoff × BackoffMultiplier
//	第3次重试: InitialBackoff × BackoffMultiplier²
//	...
//
// 示例（默认配置）:
//
//	attempt=0 → 100ms
//	attempt=1 → 200ms
//	attempt=2 → 400ms
//	attempt=3 → 800ms
//	...
//	最大不超过 MaxBackoff (5s)
func (o RetryOptions) CalculateBackoff(attempt int) time.Duration {
	backoff := o.InitialBackoff
	for i := 0; i < attempt; i++ {
		backoff = time.Duration(float64(backoff) * o.BackoffMultiplier)
		if backoff > o.MaxBackoff {
			return o.MaxBackoff
		}
	}
	return backoff
}
