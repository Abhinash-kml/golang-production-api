package middlewares

import "net/http"

func (m *MiddlewareProvider) Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := m.tracer.Start(r.Context(), "middleware.cors")
		defer span.End()

		w.Header().Set("Access-Control-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, X-Custom-Header")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Pass to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
