package observability_test

import (
	"testing"

	"github.com/kazhuravlev/lrpc/internal/observability"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/propagation"
)

func TestPropagator(t *testing.T) {
	t.Parallel()

	t.Run("case1", func(t *testing.T) {
		t.Parallel()

		res := observability.NewTextMapPropagator()
		_, ok := res.(propagation.TraceContext)
		assert.True(t, ok)
	})

	t.Run("case2", func(t *testing.T) {
		t.Parallel()

		res := observability.NewTextMapPropagator(nil, nil, nil)
		_, ok := res.(propagation.TraceContext)
		assert.True(t, ok)
	})

	t.Run("case3", func(t *testing.T) {
		t.Parallel()

		res := observability.NewTextMapPropagator(propagation.Baggage{})
		_, ok := res.(propagation.TraceContext)
		assert.False(t, ok)
		_, ok = res.(propagation.Baggage)
		assert.False(t, ok)
	})
}
