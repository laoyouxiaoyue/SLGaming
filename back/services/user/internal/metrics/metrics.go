package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	UserRegisterTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_register_total",
			Help: "Total number of user registrations",
		},
		[]string{"status"},
	)

	UserLoginTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_login_total",
			Help: "Total number of user logins",
		},
		[]string{"status", "type"},
	)

	UserLoginDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_login_duration_seconds",
			Help:    "Duration of user login operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)

	WalletRechargeTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wallet_recharge_total",
			Help: "Total number of wallet recharge operations",
		},
		[]string{"status"},
	)

	WalletRechargeAmount = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "wallet_recharge_amount",
			Help:    "Amount of wallet recharge operations",
			Buckets: []float64{10, 50, 100, 200, 500, 1000, 2000, 5000},
		},
	)

	WalletConsumeTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wallet_consume_total",
			Help: "Total number of wallet consume operations",
		},
		[]string{"status"},
	)

	WalletConsumeAmount = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "wallet_consume_amount",
			Help:    "Amount of wallet consume operations",
			Buckets: []float64{10, 50, 100, 200, 500, 1000, 2000, 5000},
		},
	)

	FollowTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_follow_total",
			Help: "Total number of follow operations",
		},
		[]string{"status", "action"},
	)

	CompanionApplyTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "companion_apply_total",
			Help: "Total number of companion applications",
		},
		[]string{"status"},
	)

	CompanionProfileUpdateTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "companion_profile_update_total",
			Help: "Total number of companion profile updates",
		},
		[]string{"status"},
	)

	RankingQueryTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ranking_query_total",
			Help: "Total number of ranking queries",
		},
		[]string{"type", "status"},
	)

	RankingQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ranking_query_duration_seconds",
			Help:    "Duration of ranking query operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)

	RedisOperationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_redis_operation_total",
			Help: "Total number of Redis operations in user service",
		},
		[]string{"operation", "status"},
	)

	DbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_db_query_duration_seconds",
			Help:    "Duration of database queries in user service",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	MqMessageTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_mq_message_total",
			Help: "Total number of MQ messages processed in user service",
		},
		[]string{"topic", "status"},
	)
)

func init() {
	prometheus.MustRegister(UserRegisterTotal)
	prometheus.MustRegister(UserLoginTotal)
	prometheus.MustRegister(UserLoginDuration)
	prometheus.MustRegister(WalletRechargeTotal)
	prometheus.MustRegister(WalletRechargeAmount)
	prometheus.MustRegister(WalletConsumeTotal)
	prometheus.MustRegister(WalletConsumeAmount)
	prometheus.MustRegister(FollowTotal)
	prometheus.MustRegister(CompanionApplyTotal)
	prometheus.MustRegister(CompanionProfileUpdateTotal)
	prometheus.MustRegister(RankingQueryTotal)
	prometheus.MustRegister(RankingQueryDuration)
	prometheus.MustRegister(RedisOperationTotal)
	prometheus.MustRegister(DbQueryDuration)
	prometheus.MustRegister(MqMessageTotal)
}
