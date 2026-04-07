package resilience

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"sync"
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
	ErrCircuitOpen    = errors.New("circuit breaker is open")
	ErrMaxRetries     = errors.New("circuit breaker max retries exceeded")
	ErrRequestTimeout = errors.New("circuit breaker request timeout")
)

// Resilient executor - type which combines Circuit breaker with retry
type Resilient struct {
	config        Config
	state         State
	failures      int
	successes     int
	lastFailure   time.Time
	halfOpenCalls int
	mu            sync.Mutex

	// Metrics
	totalRequests    int64
	successCount     int64
	failureCount     int64
	circuitOpenCount int64
}

// Constructor function
func New(config Config) *Resilient {
	return &Resilient{
		config: config,
		state:  StateClosed,
	}
}

func (r *Resilient) Execute(ctx context.Context, fn func(context.Context) error) error {
	r.mu.Lock()
	r.totalRequests++
	r.mu.Unlock()

	var lastError error

	for attempts := 1; attempts <= r.config.MaxRetries; attempts++ {
		// Check if circuit breaker will allow
		if !r.allowRequests() {
			r.mu.Lock()
			r.circuitOpenCount++
			r.mu.Lock()
			return ErrCircuitOpen
		}

		// Create timeout context for this request
		attemptCtx, cancel := context.WithTimeout(ctx, r.config.RequestTimeout)

		// Execute the function
		err := r.executeWithTimeout(attemptCtx, fn)
		cancel()

		// Record the result
		r.recordResult(err)

		// If no error then successfull call
		if err == nil {
			r.mu.Lock()
			r.successCount++
			r.mu.Unlock()
			return nil
		}

		lastError = err

		// Check if context was cancelled
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Retry only if the error is Retryable
		if !IsRetryable(err) && !errors.Is(err, ErrRequestTimeout) {
			break
		}

		// Calculate and apply exponential delay backoff
		if attempts < r.config.MaxRetries {
			delay := r.calculateDelay(attempts)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	r.mu.Lock()
	r.failureCount++
	r.mu.Unlock()

	return lastError
}

func (r *Resilient) executeWithTimeout(ctx context.Context, fn func(context.Context) error) error {
	done := make(chan error, 1)

	go func() {
		done <- fn(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ErrRequestTimeout
	}
}

func (r *Resilient) calculateDelay(attempt int) time.Duration {
	delay := float64(r.config.InittialDelay) * math.Pow(r.config.Multiplier, float64(attempt))

	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}

	if r.config.JitterFactor > 0 {
		jitter := delay * r.config.JitterFactor * (rand.Float64()*2 - 1)
		delay = delay + jitter
	}

	return time.Duration(delay)
}

func (r *Resilient) allowRequests() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch r.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(r.lastFailure) > r.config.CBTimeout {
			r.toHalfOpen()
			return r.tryHalfOpenCall()
		}
		return false
	case StateHalfOpen:
		return r.tryHalfOpenCall()
	}

	return false
}

func (r *Resilient) tryHalfOpenCall() bool {
	if r.halfOpenCalls < r.config.HalfOpenMaxCalls {
		return true
	}
	return false
}

func (r *Resilient) recordResult(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err != nil {
		r.onFailure()
	} else {
		r.onSuccess()
	}
}

func (r *Resilient) onFailure() {
	switch r.state {
	case StateClosed:
		r.failureCount++
		if r.failureCount >= int64(r.config.FailureThreshold) {
			r.toOpen()
		}
	case StateHalfOpen:
		r.toOpen()
	}

	r.lastFailure = time.Now()
}

func (r *Resilient) onSuccess() {
	switch r.state {
	case StateClosed:
		r.failures = 0
	case StateHalfOpen:
		r.halfOpenCalls--
		r.successes++
		if r.successes >= r.config.SuccessThreshold {
			r.toClosed()
		}
	}
}

func (r *Resilient) toOpen() {
	r.state = StateOpen
	r.failures = 0
	r.successes = 0
	r.halfOpenCalls = 0
}

func (r *Resilient) toHalfOpen() {
	r.state = StateHalfOpen
	r.failures = 0
	r.successes = 0
	r.halfOpenCalls = 0
}

func (r *Resilient) toClosed() {
	r.state = StateClosed
	r.failures = 0
	r.successes = 0
	r.halfOpenCalls = 0
}

func (r *Resilient) Metrics() map[string]int64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	return map[string]int64{
		"total_requests":     r.totalRequests,
		"success_count":      r.successCount,
		"failure_count":      r.failureCount,
		"circuit_open_count": r.circuitOpenCount,
	}
}

func (r *Resilient) State() State {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.state
}
