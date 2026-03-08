package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/kazhuravlev/lrpc/ctypes"
	"github.com/valyala/bytebufferpool"
)

func getIDString() ctypes.ID {
	return ctypes.ID(uuid.Must(uuid.NewV7()).String())
}

func terminateBody(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body) //nolint:errcheck
	_ = body.Close()
}

func encodePayload(buf *bytebufferpool.ByteBuffer, reqID ctypes.ID, req any) error {
	payload := Payload{
		ID:      reqID,
		Version: ctypes.Version,
		Params:  req,
	}

	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return fmt.Errorf("cannot encode payload: %w", err)
	}

	return nil
}

func parseResponse(body io.Reader, requestID ctypes.ID) (*ctypes.Response[json.RawMessage], error) {
	var resp ctypes.Response[json.RawMessage]
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Version != ctypes.Version {
		return nil, errors.New("unknown lrpc version")
	}

	if resp.ID != requestID {
		return nil, errors.New("unexpected response id")
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("service return error: %w", resp.Error)
	}

	return &resp, nil
}
