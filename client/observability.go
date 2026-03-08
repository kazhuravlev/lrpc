package client

import (
	"context"
	"net/http"

	internalobs "github.com/kazhuravlev/lrpc/internal/observability"
)

// Span represents an active trace span.
type Span = internalobs.Span

// Tracer starts spans.
type Tracer = internalobs.Tracer

// TracerProvider creates tracers.
type TracerProvider = internalobs.TracerProvider

// Propagator injects trace headers into outgoing requests.
type Propagator = interface {
	Inject(ctx context.Context, headers http.Header)
}

func defaultTracerProvider() TracerProvider {
	return internalobs.NewTracerProvider(internalobs.SpanKindClient)
}

func defaultPropagator() Propagator {
	return internalobs.NewPropagator()
}
