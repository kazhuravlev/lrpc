package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/kazhuravlev/just"
	"github.com/kazhuravlev/lrpc/ctypes"
)

// HandlerSpec stores handler internals and schema examples.
type HandlerSpec struct {
	requestSample  json.RawMessage
	responseSample json.RawMessage
	handler        handler
}

// RegisterHandler will register new handler for this method in this router.
// If you want to send concrete lrpc code on concrete errors - you can
// provide mapping "error 2 lrpc error code".
func RegisterHandler[ReqT, RespT any](
	router *Server,
	method ctypes.Method,
	innerHandler func(context.Context, ctypes.ID, ReqT) (*RespT, error),
	errorMapping map[error]ctypes.ErrorCode,
) {
	// NOTE(zhuravlev): this is need to prevent errors and race conditions,
	//   when errorMapping is changed by invoker.
	errMapping := make(map[error]ctypes.ErrorCode, len(errorMapping))
	for errorKey, errorCode := range errorMapping {
		errMapping[errorKey] = errorCode
	}

	wrapper := func(ctx context.Context, id ctypes.ID, in io.Reader) (any, *ctypes.Error) {
		var req ReqT
		if err := json.NewDecoder(in).Decode(&req); err != nil {
			router.opts.logger.ErrorContext(ctx, "decode request", slog.String("error", err.Error()))

			return nil, &ctypes.Error{
				Code:    ctypes.DefaultErrorCode,
				Message: "cannot unmarshal request",
			}
		}

		resp, err := innerHandler(ctx, id, req)
		if err != nil {
			router.opts.logger.ErrorContext(ctx, "handle request", slog.String("error", err.Error()), slog.String("id", string(id)))

			for errorKey, errorCode := range errMapping {
				if errors.Is(err, errorKey) {
					return nil, &ctypes.Error{
						Code:    errorCode,
						Message: err.Error(),
					}
				}
			}

			return nil, &ctypes.Error{
				Code:    ctypes.DefaultErrorCode,
				Message: "cannot handle request",
			}
		}

		return resp, nil
	}

	reqSample := makeSample[ReqT]()
	respSample := makeSample[RespT]()

	router.mustAddRoute(method, HandlerSpec{
		requestSample:  reqSample,
		responseSample: respSample,
		handler:        wrapper,
	})
}

// RegisterHandlerNoResponse registers a method with empty response body.
func RegisterHandlerNoResponse[ReqT any](
	router *Server,
	method ctypes.Method,
	innerHandler func(context.Context, ctypes.ID, ReqT) error,
	errorMapping map[error]ctypes.ErrorCode,
) {
	var emptyResponse noResponse

	RegisterHandler(
		router,
		method,
		func(ctx context.Context, id ctypes.ID, req ReqT) (*noResponse, error) {
			return &emptyResponse, innerHandler(ctx, id, req)
		},
		errorMapping,
	)
}

func makeSample[T any]() json.RawMessage {
	const errResult = `cannot build sample data`
	const emptyResult = `this handler has no response body`

	gofakeit.Seed(0)

	var data T
	if _, ok := any(data).(noResponse); ok {
		return just.Must(json.Marshal(emptyResult))
	}

	if err := gofakeit.Struct(&data); err != nil {
		return just.Must(json.Marshal(errResult))
	}

	dataJsonBytes, err := json.Marshal(data)
	if err != nil {
		return just.Must(json.Marshal(errResult))
	}

	return dataJsonBytes
}
