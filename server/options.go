package server

import (
	"log/slog"
)

// Options configures the server.
//
//go:generate options-gen -from-struct=Options -defaults-from=var
type Options struct {
	logger         *slog.Logger   `validate:"required"`
	ns             string         `validate:"required"`
	name           string         `validate:"required"`
	tracerProvider TracerProvider `validate:"required"`
	propagator     Propagator     `validate:"required"`
}

var defaultOptions = Options{
	logger:         slog.New(slog.DiscardHandler),
	ns:             "lrpc",
	name:           "unknown",
	tracerProvider: defaultTracerProvider(),
	propagator:     defaultPropagator(),
}
