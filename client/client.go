package client

import (
	"fmt"
	"strings"

	"github.com/kazhuravlev/lrpc/ctypes"
	"github.com/kazhuravlev/lrpc/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/valyala/bytebufferpool"
)

const instrumentationName = "observability"

var (
	mCounterVec2 = promauto.NewCounterVec(prometheus.CounterOpts{ //nolint:gochecknoglobals
		Namespace: "sdk",
		Subsystem: "lrpc_client",
		Name:      "requests_total",
		Help:      "Total count of requests that sent by lrpc client",
		ConstLabels: map[string]string{
			"lang": "go",
			"tool": "client",
		},
	}, []string{"client", "method", "status"})
	mTimeVec2 = promauto.NewHistogramVec(prometheus.HistogramOpts{ //nolint:gochecknoglobals
		Namespace: "sdk",
		Subsystem: "lrpc_client",
		Name:      "requests_duration_seconds",
		Help:      "Amount of time that lrpc client was spent on processing request",
		ConstLabels: map[string]string{
			"lang": "go",
			"tool": "client",
		},
		Buckets:                         metrics.BucketFast(),
		NativeHistogramBucketFactor:     0,
		NativeHistogramZeroThreshold:    0,
		NativeHistogramMaxBucketNumber:  0,
		NativeHistogramMinResetDuration: 0,
		NativeHistogramMaxZeroThreshold: 0,
	}, []string{"client", "method", "status"})
)

// Client sends RPC calls to a remote service.
type Client struct {
	opts Options

	bufferPool *bytebufferpool.Pool
	baseURL    string
	genID      func() ctypes.ID
	tracer     Tracer
	propagator Propagator
}

// New creates a new client.
func New(opts Options) (*Client, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options: %w", err)
	}

	tracer := opts.tracerProvider.Tracer(instrumentationName)

	return &Client{
		opts:       opts,
		bufferPool: new(bytebufferpool.Pool),
		baseURL:    strings.TrimRight(opts.serviceUrl, "/"),
		genID:      getIDString,
		tracer:     tracer,
		propagator: opts.propagator,
	}, nil
}
