package observability

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type SpanKind uint8

const (
	SpanKindClient SpanKind = iota + 1
	SpanKindServer
)

type Span interface {
	End()
}

type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, Span)
}

type TracerProvider interface {
	Tracer(name string) Tracer
}

type Propagator interface {
	Inject(ctx context.Context, headers http.Header)
	Extract(ctx context.Context, headers http.Header) context.Context
}

type tracerProvider struct {
	inner oteltrace.TracerProvider
	kind  oteltrace.SpanKind
}

func NewTracerProvider(kind SpanKind) TracerProvider {
	spanKind := oteltrace.SpanKindInternal
	switch kind {
	case SpanKindClient:
		spanKind = oteltrace.SpanKindClient
	case SpanKindServer:
		spanKind = oteltrace.SpanKindServer
	}

	return tracerProvider{
		inner: noop.NewTracerProvider(),
		kind:  spanKind,
	}
}

func (p tracerProvider) Tracer(name string) Tracer {
	return tracer{
		inner: p.inner.Tracer(name),
		kind:  p.kind,
	}
}

type tracer struct {
	inner oteltrace.Tracer
	kind  oteltrace.SpanKind
}

func (t tracer) Start(ctx context.Context, name string) (context.Context, Span) {
	ctx, span := t.inner.Start(ctx, name, oteltrace.WithSpanKind(t.kind))

	return ctx, spanAdapter{inner: span}
}

type spanAdapter struct {
	inner oteltrace.Span
}

func (s spanAdapter) End() {
	s.inner.End()
}

type propagator struct {
	inner propagation.TextMapPropagator
}

func NewPropagator() Propagator {
	return propagator{inner: NewTextMapPropagator()}
}

func (p propagator) Inject(ctx context.Context, headers http.Header) {
	p.inner.Inject(ctx, propagation.HeaderCarrier(headers))
}

func (p propagator) Extract(ctx context.Context, headers http.Header) context.Context {
	return p.inner.Extract(ctx, propagation.HeaderCarrier(headers))
}
