package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "delivery_requests_total",
			Help: "Total number of delivery requests",
		},
		[]string{"status"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "delivery_request_duration_seconds",
			Help:    "Duration of delivery request handling in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)

	DBQueryDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func ObserveRequest(status string, seconds float64) {
	RequestCount.WithLabelValues(status).Inc()
	RequestDuration.WithLabelValues(status).Observe(seconds)
}

func ObserveDBQuery(seconds float64) {
	DBQueryDuration.Observe(seconds)
}