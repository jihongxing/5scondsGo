// Package httpclient provides an HTTP client with trace ID propagation.
package httpclient

import (
	"context"
	"net/http"
	"time"

	"github.com/fiveseconds/server/pkg/trace"
)

const (
	// TraceIDHeader is the HTTP header for trace ID propagation
	TraceIDHeader = "X-Trace-ID"
)

// Client is an HTTP client that propagates trace IDs
type Client struct {
	client *http.Client
}

// New creates a new HTTP client with trace ID propagation
func New(timeout time.Duration) *Client {
	return &Client{
		client: &http.Client{
			Timeout: timeout,
			Transport: &tracingTransport{
				base: http.DefaultTransport,
			},
		},
	}
}

// Do executes an HTTP request with trace ID propagation
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// Get performs a GET request with trace ID propagation
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// tracingTransport is an http.RoundTripper that adds trace ID to requests
type tracingTransport struct {
	base http.RoundTripper
}

func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Get trace ID from context and add to header
	if traceID := trace.GetTraceID(req.Context()); traceID != "" {
		req.Header.Set(TraceIDHeader, traceID)
	}
	return t.base.RoundTrip(req)
}
