package httpclient

import (
	"net/http"
	"time"

	"github.com/abhinash-kml/go-api-server/internal/resilience"
)

// Client wraps default http client with resilience patterns
type client struct {
	httpclient *http.Client
	resilient  *resilience.Resilient
}

// Constructor function
func NewClient(config resilience.Config) *client {
	return &client{
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
