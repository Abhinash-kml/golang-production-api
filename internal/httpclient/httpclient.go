package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/abhinash-kml/go-api-server/internal/resilience"
)

// Client wraps default http client with resilience patterns
type Client struct {
	httpclient *http.Client
	resilient  *resilience.Resilient
}

// Constructor function
func NewClient(config resilience.Config) *Client {
	return &Client{
		httpclient: &http.Client{
			Timeout: config.RequestTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		resilient: resilience.New(config),
	}
}

// Response wraps the HTTP response
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// Get performs a resilient GET request
func (c *Client) Get(ctx context.Context, url string) (*Response, error) {
	var response *Response

	err := c.resilient.Execute(ctx, func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		resp, err := c.httpclient.Do(req)
		if err != nil {
			// Network errors are retryable
			return resilience.RetryableError{Err: err}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return resilience.RetryableError{Err: err}
		}

		// 5xx errors are retryable
		if resp.StatusCode >= 500 {
			return resilience.RetryableError{
				Err: fmt.Errorf("server error: %d", resp.StatusCode),
			}
		}

		// 4xx errors are not retryable
		if resp.StatusCode >= 400 {
			return fmt.Errorf("client error: %d", resp.StatusCode)
		}

		response = &Response{
			StatusCode: resp.StatusCode,
			Body:       body,
			Headers:    resp.Header,
		}

		return nil
	})

	return response, err
}

// Metrics
func (c *Client) Metrics() map[string]int64 {
	return c.resilient.Metrics()
}
