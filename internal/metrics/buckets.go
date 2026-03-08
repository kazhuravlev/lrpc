package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func BucketFast() []float64 {
	return prometheus.ExponentialBuckets(0.001, 2.5, 8) //nolint:gomnd
}

func BucketMedium() []float64 {
	return prometheus.ExponentialBuckets(0.01, 2, 8) //nolint:gomnd
}

func BucketSlow() []float64 {
	return prometheus.ExponentialBuckets(0.1, 2, 6) //nolint:gomnd
}
