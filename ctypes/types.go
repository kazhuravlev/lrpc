package ctypes

import (
	"encoding/json"
	"fmt"
)

const Version = "1"

// ErrorCode is a numeric service error code.
type ErrorCode int

const DefaultErrorCode ErrorCode = 666

// Method is an RPC method name.
type Method string

// ID is a request identifier.
type ID string

// Error is an RPC error payload.
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// Error returns a readable error string.
func (e Error) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// Request is a common RPC request body.
type Request struct {
	Version string          `json:"lrpc"` //nolint:tagliatelle
	ID      ID              `json:"id"`
	Params  json.RawMessage `json:"params"`
}

// Response is a common RPC response body.
type Response[T any] struct {
	Version string `json:"lrpc"` //nolint:tagliatelle
	ID      ID     `json:"id"`
	Error   *Error `json:"error,omitempty"`
	Result  T      `json:"result,omitempty"`
}
