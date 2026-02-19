package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// DriftHistogram measures the delay between ScheduledTime and Actual Execution Time.
	// This is the critical metric for scheduler accuracy.
	DriftHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "scheduler_task_drift_seconds",
		Help:    "Difference between Scheduled Time and Actual Execution Time in seconds",
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10, 60, 300}, // 1ms to 5m
	}, []string{"task_type"})

	// LatencyHistogram measures the duration of the task execution itself (worker processing time).
	LatencyHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "scheduler_task_execution_duration_seconds",
		Help:    "Duration of task execution in seconds",
		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10, 30, 60},
	}, []string{"task_type", "status"})

	// TaskCounter counts the number of tasks processed.
	TaskCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "scheduler_tasks_total",
		Help: "Total number of tasks processed by the scheduler",
	}, []string{"task_type", "status"}) // status: success, failed, retrying

	// QueueDepth gauge tracks the number of tasks currently in the wheel's memory.
	QueueDepth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "scheduler_wheel_queue_depth",
		Help: "Number of tasks currently in the timing wheel memory",
	}, []string{"bucket_type"}) // bucket_type: seconds, minutes, hours, unprocessed
)
