package middlewares

import (
	"net/http"

	"go.uber.org/zap"
)

type MiddleWare func(http.Handler) http.Handler

func CompileHandlers(base http.Handler, middlewares ...MiddleWare) http.Handler {
	final := base

	for i := len(middlewares) - 1; i >= 0; i-- {
		final = middlewares[i](final)
	}

	return final
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zap.L().Info("Connection", zap.String("IP", r.RemoteAddr), zap.String("Method", r.Method), zap.String("Path", r.Pattern))

		next.ServeHTTP(w, r)
	})
}
