package metrics

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func VecMethodExecCount(namespace, mod, name string) *prometheus.CounterVec {
	metricName := fmt.Sprintf("%s_%s_exec_count", mod, name)

	return prometheus.NewCounterVec(prometheus.CounterOpts{ //nolint:exhaustruct
		Namespace: namespace,
		Name:      metricName,
		Help:      fmt.Sprintf("Counter of execution count of %s:%s:%s methods.", namespace, mod, name),
	}, []string{"method", "status"})
}

func VecMethodExecDuration(namespace, mod, name string, buckets []float64) *prometheus.HistogramVec {
	metricName := fmt.Sprintf("%s_%s_exec_time_seconds", mod, name)

	return prometheus.NewHistogramVec(prometheus.HistogramOpts{ //nolint:exhaustruct
		Namespace: namespace,
		Name:      metricName,
		Help:      fmt.Sprintf("Histogram of execution time of %s:%s:%s (seconds).", namespace, mod, name),
		Buckets:   buckets,
	}, []string{"method", "status"})
}

func VecGauge(namespace, mod, name string) *prometheus.GaugeVec {
	metricName := fmt.Sprintf("%s_%s_count", mod, name)

	return prometheus.NewGaugeVec(prometheus.GaugeOpts{ //nolint:exhaustruct
		Namespace: namespace,
		Name:      metricName,
		Help:      fmt.Sprintf("Gauge for count of %s:%s:%s.", namespace, mod, name),
	}, []string{"tag"})
}

func RegisterCollector[T prometheus.Collector](registerer prometheus.Registerer, collector T) (T, error) {
	if err := registerer.Register(collector); err != nil {
		var alreadyRegisteredErr prometheus.AlreadyRegisteredError
		if errors.As(err, &alreadyRegisteredErr) {
			existing, ok := alreadyRegisteredErr.ExistingCollector.(T)
			if !ok {
				var zero T

				return zero, fmt.Errorf("existing collector has unexpected type: %T", alreadyRegisteredErr.ExistingCollector)
			}

			return existing, nil
		}

		var zero T

		return zero, fmt.Errorf("register collector: %w", err)
	}

	return collector, nil
}
