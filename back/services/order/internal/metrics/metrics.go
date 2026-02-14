package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	OrderCreateTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_create_total",
			Help: "Total number of order creation operations",
		},
		[]string{"status"},
	)

	OrderCreateDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_create_duration_seconds",
			Help:    "Duration of order creation operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{},
	)

	OrderCancelTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_cancel_total",
			Help: "Total number of order cancellation operations",
		},
		[]string{"status", "need_refund"},
	)

	OrderCancelDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_cancel_duration_seconds",
			Help:    "Duration of order cancellation operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{},
	)

	OrderAcceptTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_accept_total",
			Help: "Total number of order acceptance operations",
		},
		[]string{"status"},
	)

	OrderAcceptDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_accept_duration_seconds",
			Help:    "Duration of order acceptance operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{},
	)

	OrderCompleteTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_complete_total",
			Help: "Total number of order completion operations",
		},
		[]string{"status"},
	)

	OrderCompleteDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_complete_duration_seconds",
			Help:    "Duration of order completion operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{},
	)

	OrderAmount = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "order_amount",
			Help:    "Amount of orders in coins",
			Buckets: []float64{10, 50, 100, 200, 500, 1000, 2000, 5000},
		},
	)

	OrderStatusTransition = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_status_transition_total",
			Help: "Total number of order status transitions",
		},
		[]string{"from_status", "to_status"},
	)

	DbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_db_query_duration_seconds",
			Help:    "Duration of database queries in order service",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	MqMessageTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_mq_message_total",
			Help: "Total number of MQ messages processed in order service",
		},
		[]string{"topic", "event_type", "status"},
	)

	MqMessageDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_mq_message_duration_seconds",
			Help:    "Duration of MQ message processing in order service",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topic", "event_type"},
	)

	LockAcquireTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_lock_acquire_total",
			Help: "Total number of distributed lock acquisitions in order service",
		},
		[]string{"lock_type", "status"},
	)

	LockAcquireDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_lock_acquire_duration_seconds",
			Help:    "Duration of distributed lock acquisition in order service",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"lock_type"},
	)
)

func init() {
	prometheus.MustRegister(OrderCreateTotal)
	prometheus.MustRegister(OrderCreateDuration)
	prometheus.MustRegister(OrderCancelTotal)
	prometheus.MustRegister(OrderCancelDuration)
	prometheus.MustRegister(OrderAcceptTotal)
	prometheus.MustRegister(OrderAcceptDuration)
	prometheus.MustRegister(OrderCompleteTotal)
	prometheus.MustRegister(OrderCompleteDuration)
	prometheus.MustRegister(OrderAmount)
	prometheus.MustRegister(OrderStatusTransition)
	prometheus.MustRegister(DbQueryDuration)
	prometheus.MustRegister(MqMessageTotal)
	prometheus.MustRegister(MqMessageDuration)
	prometheus.MustRegister(LockAcquireTotal)
	prometheus.MustRegister(LockAcquireDuration)
}
