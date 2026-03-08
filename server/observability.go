package server

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

// Propagator extracts trace headers from incoming requests.
type Propagator = interface {
	Extract(ctx context.Context, headers http.Header) context.Context
}

func defaultTracerProvider() TracerProvider {
	return internalobs.NewTracerProvider(internalobs.SpanKindServer)
}

func defaultPropagator() Propagator {
	return internalobs.NewPropagator()
}
