package middlewares

import (
	"net/http"

	"go.uber.org/zap"
)

func (m *MiddlewareProvider) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := m.tracer.Start(r.Context(), "middleare.Logger")
		defer span.End()

		zap.L().Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
