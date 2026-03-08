package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kazhuravlev/lrpc/ctypes"
)

const (
	mStatusErr = "err"
	mStatusOk  = "ok"
)

// RoundTrip sends one RPC request and decodes response into respObj.
func (c *Client) RoundTrip(ctx context.Context, method ctypes.Method, req any, respObj any) error {
	start := time.Now()

	ctx, span := c.tracer.Start(ctx, string(method))
	defer span.End()

	requestID := c.genID()

	buf := c.bufferPool.Get()
	defer c.bufferPool.Put(buf)

	// Encode payload
	if err := encodePayload(buf, requestID, req); err != nil {
		c.collectMetrics(string(method), mStatusErr, start)

		return fmt.Errorf("encode payload: %w", err)
	}

	// Send request/parse response
	response, err := c.doRequest(ctx, method, requestID, bytes.NewReader(buf.Bytes()))
	if err != nil {
		c.collectMetrics(string(method), mStatusErr, start)

		return fmt.Errorf("do request: %w", err)
	}

	if respObj != nil {
		if err := json.Unmarshal(response.Result, respObj); err != nil {
			c.collectMetrics(string(method), mStatusErr, start)

			return fmt.Errorf("unmarshal response payload: %w", err)
		}
	}

	c.collectMetrics(string(method), mStatusOk, start)

	return nil
}
