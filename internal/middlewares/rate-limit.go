package middlewares

import (
	"net/http"
	"time"

	"github.com/abhinash-kml/go-api-server/pkg/ratelimiter"
	"go.opentelemetry.io/otel/attribute"
)

var limiter = ratelimiter.FixedWindowLimiter{
	WindowDuration: time.Second * 10,
	LimitPerWindow: 5,
	Table:          make(map[string]*ratelimiter.ClientInfo),
}

func (m *MiddlewareProvider) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := m.tracer.Start(r.Context(), "middleware.RateLimit")
		defer span.End()

		// Rate limit logic
		if !limiter.Allow(r.RemoteAddr) {
			span.SetAttributes(attribute.Bool("ratelimited", true))
			http.Error(w, "Rate Limited", http.StatusTooManyRequests)
			return
		}

		span.SetAttributes(attribute.Bool("ratelimited", false))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
