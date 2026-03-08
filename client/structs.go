package client

import "github.com/kazhuravlev/lrpc/ctypes"

// Payload is the request body sent by client.
type Payload struct {
	ID      ctypes.ID `json:"id"`
	Version string    `json:"lrpc"` //nolint:tagliatelle
	Params  any       `json:"params"`
}
