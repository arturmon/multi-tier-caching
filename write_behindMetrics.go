package multi_tier_caching

import "github.com/prometheus/client_golang/prometheus"

var (
	queueLengthGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "write_queue_length",
			Help: "Current number of tasks in the write queue",
		},
	)

	processedTasksCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "write_queue_processed_total",
			Help: "Total number of processed tasks in the write queue",
		},
	)

	taskProcessingHistogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "write_queue_processing_time_seconds",
			Help:    "Histogram of write task processing times",
			Buckets: prometheus.LinearBuckets(0.01, 0.05, 10), // от 10ms до 500ms
		},
	)
)
