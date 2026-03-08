package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/kazhuravlev/lrpc/ctypes"
)

func (c *Client) writeHeaders(req *http.Request) {
	req.Header.Set("content-type", "application/json")
	if c.opts.serviceToken != "" {
		req.Header.Set("authorization", "Bearer "+c.opts.serviceToken)
	}
}

func (c *Client) collectMetrics(method, status string, start time.Time) {
	mCounterVec2.WithLabelValues(c.opts.name, method, status).Inc()
	mTimeVec2.WithLabelValues(c.opts.name, method, status).Observe(time.Since(start).Seconds())
}

func (c *Client) requestURL(method ctypes.Method) string {
	return c.baseURL + "/" + url.PathEscape(string(method))
}

func (c *Client) doRequest(ctx context.Context, method ctypes.Method, requestID ctypes.ID, buf io.Reader) (*ctypes.Response[json.RawMessage], error) {
	httpReq, err := http.NewRequest(http.MethodPost, c.requestURL(method), buf)
	if err != nil {
		return nil, fmt.Errorf("cannot create new request: %w", err)
	}

	c.writeHeaders(httpReq)

	c.propagator.Inject(ctx, httpReq.Header)

	resp, err := c.opts.httpClient.Do(httpReq.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("cannot do request: %w", err)
	}
	defer terminateBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		c.opts.logger.ErrorContext(ctx, "unexpected status code",
			slog.Int("status", resp.StatusCode))

		return nil, errors.New("unexpected status code") //nolint:goerr113 // this should not be handled by client.
	}

	// Parse response
	response, err := parseResponse(resp.Body, requestID)
	if err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return response, nil
}
