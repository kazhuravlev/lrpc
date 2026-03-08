package metrics_test

import (
	"testing"
	"time"

	"github.com/kazhuravlev/lrpc/internal/metrics"
)

//nolint:gochecknoglobals // avoid problems with benchmark mechanism
var (
	counterVec = metrics.VecMethodExecCount("namespace", "module", "name")
	gaugeVec   = metrics.VecGauge("namespace", "module", "name")
	histVec    = metrics.VecMethodExecDuration("namespace", "module", "name", metrics.BucketFast())
)

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
