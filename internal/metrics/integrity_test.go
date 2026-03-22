package metrics_test

import (
	"testing"
	"time"

	"github.com/kazhuravlev/lrpc/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

//nolint:gochecknoglobals // avoid problems with benchmark mechanism
var (
	registry   = prometheus.NewRegistry()
	counterVec = must(metrics.RegisterCollector(registry, metrics.VecMethodExecCount("namespace", "module", "name")))
	gaugeVec   = must(metrics.RegisterCollector(registry, metrics.VecGauge("namespace", "module", "name")))
	histVec    = must(metrics.RegisterCollector(registry, metrics.VecMethodExecDuration("namespace", "module", "name", metrics.BucketFast())))
)

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}

	return val
}

func BenchmarkPrometheusCounterVec(b *testing.B) {
	counterVec.Reset()
	for i := 0; i < b.N; i++ {
		counterVec.WithLabelValues("get-some-data-from-some-service", "ok").Inc()
	}
}

func BenchmarkPrometheusGaugeVec(b *testing.B) {
	gaugeVec.Reset()
	for i := 0; i < b.N; i++ {
		gaugeVec.WithLabelValues("get-some-data-from-some-service").Inc()
	}
}

func BenchmarkPrometheusHistVec(b *testing.B) {
	values := []float64{
		(10 * time.Millisecond).Seconds(),
		(100 * time.Millisecond).Seconds(),
		(500 * time.Millisecond).Seconds(),
		(1 * time.Second).Seconds(),
	}

	histVec.Reset()
	for i := 0; i < b.N; i++ {
		histVec.WithLabelValues("get-some-data-from-some-service", "ok").Observe(values[i%len(values)])
	}
}
