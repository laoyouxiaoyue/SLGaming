package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// CodeSendTotal 验证码发送总数
	CodeSendTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "code_send_total",
			Help: "Total number of verification codes sent",
		},
		[]string{"purpose", "status"},
	)

	// CodeSendDuration 验证码发送耗时
	CodeSendDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "code_send_duration_seconds",
			Help:    "Duration of code send operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"purpose"},
	)

	// CodeVerifyTotal 验证码验证总数
	CodeVerifyTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "code_verify_total",
			Help: "Total number of verification code verifications",
		},
		[]string{"status"},
	)

	// CodeVerifyDuration 验证码验证耗时
	CodeVerifyDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "code_verify_duration_seconds",
			Help:    "Duration of code verify operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// CodeRateLimitTotal 限流拦截总数
	CodeRateLimitTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "code_ratelimit_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"type"},
	)

	// CodeRedisErrorTotal Redis 错误总数
	CodeRedisErrorTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "code_redis_error_total",
			Help: "Total number of Redis operation errors",
		},
	)
)

func init() {
	prometheus.MustRegister(CodeSendTotal)
	prometheus.MustRegister(CodeSendDuration)
	prometheus.MustRegister(CodeVerifyTotal)
	prometheus.MustRegister(CodeVerifyDuration)
	prometheus.MustRegister(CodeRateLimitTotal)
	prometheus.MustRegister(CodeRedisErrorTotal)
}
