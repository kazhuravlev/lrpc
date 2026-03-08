package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/kazhuravlev/lrpc/ctypes"
)

type benchReq struct {
	Value string `json:"value"`
}

type benchResp struct {
	Value string `json:"value"`
}

var benchServerSeq uint64

func nextBenchServerName(prefix string) string {
	n := atomic.AddUint64(&benchServerSeq, 1)

	return fmt.Sprintf("%s_%d", prefix, n)
}

func benchmarkRequestBody(b *testing.B, req ctypes.Request) []byte {
	b.Helper()

	raw, err := json.Marshal(req)
	if err != nil {
		b.Fatalf("marshal request: %v", err)
	}

	return raw
}

func BenchmarkHTTPHandlerSuccess(b *testing.B) {
	srv, err := New(NewOptions(WithName(nextBenchServerName("bench_http_handler_success"))))
	if err != nil {
		b.Fatalf("new server: %v", err)
	}

	RegisterHandler(
		srv,
		"bench-echo-v1",
		func(_ context.Context, _ ctypes.ID, req benchReq) (*benchResp, error) {
			return &benchResp{Value: req.Value}, nil
		},
		nil,
	)

	handler := srv.HTTPHandler()
	body := benchmarkRequestBody(b, ctypes.Request{
		Version: ctypes.Version,
		ID:      "bench-id",
		Params:  json.RawMessage(`{"value":"hello"}`),
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/lrpc/bench-echo-v1", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", w.Code)
		}
	}
}

func BenchmarkHTTPHandlerMappedError(b *testing.B) {
	errBusiness := errors.New("business error")

	srv, err := New(NewOptions(WithName(nextBenchServerName("bench_http_handler_error"))))
	if err != nil {
		b.Fatalf("new server: %v", err)
	}

	RegisterHandler(
		srv,
		"bench-echo-v1",
		func(_ context.Context, _ ctypes.ID, _ benchReq) (*benchResp, error) {
			return nil, errBusiness
		},
		map[error]ctypes.ErrorCode{
			errBusiness: 1001,
		},
	)

	handler := srv.HTTPHandler()
	body := benchmarkRequestBody(b, ctypes.Request{
		Version: ctypes.Version,
		ID:      "bench-id",
		Params:  json.RawMessage(`{"value":"hello"}`),
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/lrpc/bench-echo-v1", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", w.Code)
		}
	}
}

func BenchmarkHTTPHandlerBadRequest(b *testing.B) {
	srv, err := New(NewOptions(WithName(nextBenchServerName("bench_http_handler_bad_request"))))
	if err != nil {
		b.Fatalf("new server: %v", err)
	}

	handler := srv.HTTPHandler()
	body := []byte("{")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/lrpc/bench-echo-v1", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			b.Fatalf("unexpected status: %d", w.Code)
		}
	}
}
