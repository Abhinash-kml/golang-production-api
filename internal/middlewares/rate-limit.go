package middlewares

import (
	"net/http"
	"time"

	"github.com/abhinash-kml/go-api-server/pkg/ratelimiter"
)

var limiter = ratelimiter.FixedWindowLimiter{
	WindowDuration: time.Second * 10,
	LimitPerWindow: 5,
	Table:          make(map[string]*ratelimiter.ClientInfo),
}

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Rate limit logic
		if !limiter.Allow(r.RemoteAddr) {
			http.Error(w, "Rate Limited", http.StatusTooManyRequests)
		}

		next.ServeHTTP(w, r)
	})
}
