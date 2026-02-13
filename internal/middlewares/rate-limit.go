package middlewares

import "net/http"

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Rate limit logic

		next.ServeHTTP(w, r)
	})
}
