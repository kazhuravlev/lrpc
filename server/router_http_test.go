package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kazhuravlev/just"
	"github.com/kazhuravlev/lrpc/ctypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHTTPRequest(t *testing.T) {
	t.Parallel()

	mkBody := func(version string) []byte {
		t.Helper()
		raw, err := json.Marshal(ctypes.Request{
			Version: version,
			ID:      "req-1",
			Params:  json.RawMessage(`{"x":1}`),
		})
		require.NoError(t, err)

		return raw
	}

	t.Run("ok_with_path_value", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/lrpc/ignored", bytes.NewReader(mkBody(ctypes.Version)))
		req.SetPathValue("method", "sum-v1")

		parsed, method, err := parseHTTPRequest(req)
		require.NoError(t, err)
		assert.Equal(t, ctypes.Method("sum-v1"), method)
		assert.Equal(t, ctypes.ID("req-1"), parsed.ID)
	})

	t.Run("ok_with_path_fallback", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/lrpc/sum-v2", bytes.NewReader(mkBody(ctypes.Version)))

		_, method, err := parseHTTPRequest(req)
		require.NoError(t, err)
		assert.Equal(t, ctypes.Method("sum-v2"), method)
	})

	t.Run("bad_method", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(mkBody(ctypes.Version)))

		_, _, err := parseHTTPRequest(req)
		require.Error(t, err)
	})

	t.Run("bad_json", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/lrpc/sum-v1", bytes.NewReader([]byte("{")))

		_, _, err := parseHTTPRequest(req)
		require.Error(t, err)
	})

	t.Run("bad_version", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/lrpc/sum-v1", bytes.NewReader(mkBody("bad-version")))

		_, _, err := parseHTTPRequest(req)
		require.Error(t, err)
	})
}

func TestHTTPHandler(t *testing.T) {
	t.Parallel()

	type EchoReq struct {
		Value string `json:"value"`
	}
	type EchoResp struct {
		Value string `json:"value"`
	}

	errBusiness := errors.New("business error")

	srv := just.Must(New(NewOptions(WithName("router_http_test_server"))))
	RegisterHandler(
		srv,
		"echo-v1",
		func(_ context.Context, _ ctypes.ID, req EchoReq) (*EchoResp, error) {
			if req.Value == "err" {
				return nil, errBusiness
			}

			return &EchoResp{Value: req.Value}, nil
		},
		map[error]ctypes.ErrorCode{errBusiness: 1001},
	)

	handler := srv.HTTPHandler()

	call := func(t *testing.T, path string, body any) (int, ctypes.Response[json.RawMessage]) {
		t.Helper()

		rawBody, err := json.Marshal(body)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(rawBody))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var resp ctypes.Response[json.RawMessage]
		if w.Code == http.StatusOK {
			err = json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
		}

		return w.Code, resp
	}

	t.Run("ok", func(t *testing.T) {
		t.Parallel()

		status, resp := call(t, "/api/v1/lrpc/echo-v1", ctypes.Request{
			Version: ctypes.Version,
			ID:      "id-ok",
			Params:  json.RawMessage(`{"value":"hello"}`),
		})
		require.Equal(t, http.StatusOK, status)
		require.Nil(t, resp.Error)
		assert.Equal(t, ctypes.ID("id-ok"), resp.ID)

		var payload EchoResp
		require.NoError(t, json.Unmarshal(resp.Result, &payload))
		assert.Equal(t, "hello", payload.Value)
	})

	t.Run("mapped_error", func(t *testing.T) {
		t.Parallel()

		status, resp := call(t, "/api/v1/lrpc/echo-v1", ctypes.Request{
			Version: ctypes.Version,
			ID:      "id-err",
			Params:  json.RawMessage(`{"value":"err"}`),
		})
		require.Equal(t, http.StatusOK, status)
		require.NotNil(t, resp.Error)
		assert.Equal(t, ctypes.ErrorCode(1001), resp.Error.Code)
	})

	t.Run("unknown_method", func(t *testing.T) {
		t.Parallel()

		status, resp := call(t, "/api/v1/lrpc/missing-v1", ctypes.Request{
			Version: ctypes.Version,
			ID:      "id-missing",
			Params:  json.RawMessage(`{}`),
		})
		require.Equal(t, http.StatusOK, status)
		require.NotNil(t, resp.Error)
		assert.Equal(t, ctypes.DefaultErrorCode, resp.Error.Code)
		assert.Equal(t, "unknown method", resp.Error.Message)
	})

	t.Run("bad_request_status", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/lrpc/echo-v1", bytes.NewReader([]byte("{")))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
