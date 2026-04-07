package resilience

import (
	"errors"
	"time"
)

// Retryable error is an error where operation should be retried
type RetryableError struct {
	Err error
}

func (r RetryableError) Error() string {
	return r.Err.Error()
}

func (r RetryableError) Unwrap() error {
	return r.Err
}

func IsRetryable(err error) bool {
	var retryable RetryableError
	return errors.As(err, &retryable)
}

type Config struct {
	// Retry settings
	MaxRetries    int
	InittialDelay time.Duration
	MaxDelay      time.Duration
	Multiplier    float64
	JitterFactor  float64

	// Circuit breaker settings
	FailureThreshold int
	SuccessThreshold int
	CBTimeout        time.Duration
	HalfOpenMaxCalls int

	// General settings
	RequestTimeout time.Duration
}

func DefaultConfig() Config {
	return Config{
		MaxRetries:       3,
		InittialDelay:    100 * time.Millisecond,
		MaxDelay:         10 * time.Second,
		Multiplier:       2.0,
		JitterFactor:     0.2,
		FailureThreshold: 5,
		SuccessThreshold: 2,
		CBTimeout:        30 * time.Second,
		HalfOpenMaxCalls: 1,
		RequestTimeout:   5 * time.Second,
	}
}

// Circuit breaker state represented as int
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var (
	ErrStateOpen      = errors.New("circuit breaker is open")
	ErrMaxRetries     = errors.New("circuit breaker max retries exceeded")
	ErrRequestTimeout = errors.New("circuit breaker request timeout")
)
