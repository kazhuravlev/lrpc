package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func VecMethodExecCount(namespace, mod, name string) *prometheus.CounterVec {
	metricName := fmt.Sprintf("%s_%s_exec_count", mod, name)

	return promauto.NewCounterVec(prometheus.CounterOpts{ //nolint:exhaustruct
		Namespace: namespace,
		Name:      metricName,
		Help:      fmt.Sprintf("Counter of execution count of %s:%s:%s methods.", namespace, mod, name),
	}, []string{"method", "status"})
}

func VecMethodExecDuration(namespace, mod, name string, buckets []float64) *prometheus.HistogramVec {
	metricName := fmt.Sprintf("%s_%s_exec_time_seconds", mod, name)

	return promauto.NewHistogramVec(prometheus.HistogramOpts{ //nolint:exhaustruct
		Namespace: namespace,
		Name:      metricName,
		Help:      fmt.Sprintf("Histogram of execution time of %s:%s:%s (seconds).", namespace, mod, name),
		Buckets:   buckets,
	}, []string{"method", "status"})
}

func VecGauge(namespace, mod, name string) *prometheus.GaugeVec {
	metricName := fmt.Sprintf("%s_%s_count", mod, name)

	return promauto.NewGaugeVec(prometheus.GaugeOpts{ //nolint:exhaustruct
		Namespace: namespace,
		Name:      metricName,
		Help:      fmt.Sprintf("Gauge for count of %s:%s:%s.", namespace, mod, name),
	}, []string{"tag"})
}
