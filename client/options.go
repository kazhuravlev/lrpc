package client

import (
	"log/slog"
	"net/http"
)

// Options configures the client.
//
//go:generate options-gen -from-struct=Options -defaults-from=func
type Options struct {
	logger         *slog.Logger `validate:"required"`
	serviceUrl     string       `validate:"required,url"`
	httpClient     IHttpClient  `validate:"required"`
	serviceToken   string
	tracerProvider TracerProvider `validate:"required"`
	propagator     Propagator     `validate:"required"`
	// The client name that used by instrumentation.
	name string `validate:"required"`
}

func getDefaultOptions() Options {
	return Options{
		logger:         slog.New(slog.DiscardHandler),
		serviceUrl:     "http://127.0.0.1:8000/api/v1/lrpc",
		httpClient:     http.DefaultClient,
		serviceToken:   "",
		tracerProvider: defaultTracerProvider(),
		propagator:     defaultPropagator(),
		name:           "some_client",
	}
}

// IHttpClient is a minimal HTTP client interface.
type IHttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}
