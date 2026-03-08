package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/kazhuravlev/lrpc/ctypes"
	"github.com/kazhuravlev/lrpc/internal/metrics"

	"github.com/kazhuravlev/just"
	"github.com/prometheus/client_golang/prometheus"
)

const instrumentationName = "observability"

type handler func(context.Context, ctypes.ID, io.Reader) (any, *ctypes.Error)

// Server handles RPC requests.
type Server struct {
	opts Options

	routes     map[ctypes.Method]HandlerSpec
	tracer     Tracer
	propagator Propagator

	timeVec    *prometheus.HistogramVec
	counterVec *prometheus.CounterVec
}

// New creates a new server.
func New(opts Options) (*Server, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options: %w", err)
	}

	tracer := opts.tracerProvider.Tracer(instrumentationName)

	return &Server{
		opts:       opts,
		routes:     make(map[ctypes.Method]HandlerSpec),
		tracer:     tracer,
		propagator: opts.propagator,
		timeVec:    metrics.VecMethodExecDuration(opts.ns, opts.name, "requests", metrics.BucketFast()),
		counterVec: metrics.VecMethodExecCount(opts.ns, opts.name, "requests"),
	}, nil
}

func (r *Server) mustAddRoute(method ctypes.Method, handlerSpec HandlerSpec) {
	if _, ok := r.routes[method]; ok {
		panic("method is already registered")
	}

	r.routes[method] = handlerSpec
}

func (r *Server) observeMetrics(method, status string, start time.Time) {
	r.timeVec.
		WithLabelValues(method, status).
		Observe(time.Since(start).Seconds())
	r.counterVec.
		WithLabelValues(method, status).
		Inc()
}

// HTTPHandler returns an HTTP handler for lrpc requests.
func (r *Server) HTTPHandler() http.Handler { //nolint:funlen
	const mStatusOk = "ok"
	const mStatusErr = "err"
	const mUnknownMethod = "__unknown__"

	return http.HandlerFunc(func(w http.ResponseWriter, reqHTTP *http.Request) {
		start := time.Now()
		req, method, err := parseHTTPRequest(reqHTTP)
		if err != nil {
			r.opts.logger.ErrorContext(reqHTTP.Context(), "bad request", slog.String("error", err.Error()))
			r.observeMetrics(mUnknownMethod, mStatusErr, start)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		extractedCtx := r.propagator.Extract(reqHTTP.Context(), reqHTTP.Header)

		ctx, span := r.tracer.Start(extractedCtx, string(method))
		defer span.End()

		metricMethod := string(method)
		handler, ok := r.routes[method]
		if !ok {
			// NOTE(zhuravlev): this is an important thing because attackers can make a flood and blow metrics cardinality.
			metricMethod = mUnknownMethod
		}

		var res any
		var err2 *ctypes.Error
		if !ok {
			res = nil
			err2 = &ctypes.Error{
				Code:    ctypes.DefaultErrorCode,
				Message: "unknown method",
			}
		} else {
			res, err2 = handler.handler(ctx, req.ID, bytes.NewReader(req.Params))
		}

		if err2 != nil {
			r.observeMetrics(metricMethod, mStatusErr, start)
			writeJSON(w, http.StatusOK, ctypes.Response[any]{
				Version: ctypes.Version,
				ID:      req.ID,
				Error:   err2,
				Result:  nil,
			})

			return
		}

		r.observeMetrics(metricMethod, mStatusOk, start)
		writeJSON(w, http.StatusOK, ctypes.Response[any]{
			Version: ctypes.Version,
			ID:      req.ID,
			Error:   nil,
			Result:  res,
		})
	})
}

// HTTPHandlerSchema returns a handler with methods schema.
func (r *Server) HTTPHandlerSchema() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		schema := r.getSchemaJSON(req.Context())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(schema)
	})
}

// Struct keeps one schema example.
type Struct struct {
	Example json.RawMessage `json:"example"`
}

// HandlerSchema describes one registered method.
type HandlerSchema struct {
	Method   string `json:"method"`
	Request  Struct `json:"request"`
	Response Struct `json:"response"`
}

// getSchemaJSON returns schema for all registered methods.
func (r *Server) getSchemaJSON(ctx context.Context) json.RawMessage {
	methodNames := just.SliceMap(
		just.MapGetKeys(r.routes),
		func(method ctypes.Method) string {
			return string(method)
		},
	)

	sort.Strings(methodNames)

	handlerSchemas := just.SliceMap(
		just.MapPairs(r.routes),
		func(methHandler just.KV[ctypes.Method, HandlerSpec]) HandlerSchema {
			return HandlerSchema{
				Method: string(methHandler.Key),
				Request: Struct{
					Example: methHandler.Val.requestSample,
				},
				Response: Struct{
					Example: methHandler.Val.responseSample,
				},
			}
		},
	)

	just.SliceSort(handlerSchemas, func(a, b HandlerSchema) bool { return a.Method < b.Method })

	schemaBytes, err := json.MarshalIndent(
		map[string]any{"handlers": handlerSchemas},
		"",
		" ",
	)
	if err != nil {
		r.opts.logger.WarnContext(ctx, "marshal schema", slog.String("error", err.Error()))
	}

	return schemaBytes
}

func parseHTTPRequest(reqHTTP *http.Request) (*ctypes.Request, ctypes.Method, error) {
	method := ctypes.Method(reqHTTP.PathValue("method"))
	if method == "" {
		path := strings.TrimRight(reqHTTP.URL.Path, "/")
		idx := strings.LastIndexByte(path, '/')
		if idx < 0 || idx == len(path)-1 {
			return nil, "", errors.New("bad method")
		}

		method = ctypes.Method(path[idx+1:])
	}

	if method == "" {
		return nil, "", errors.New("bad method")
	}

	var req ctypes.Request
	if err := json.NewDecoder(reqHTTP.Body).Decode(&req); err != nil {
		return nil, "", fmt.Errorf("bad request: %w", err)
	}

	if req.Version != ctypes.Version {
		return nil, "", errors.New("bad version")
	}

	return &req, method, nil
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	raw, err := json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot marshal response: %v", err), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(raw)
}
