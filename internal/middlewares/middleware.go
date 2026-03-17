package middlewares

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

type MiddlewareProvider struct {
	tracer trace.Tracer
}

func NewMiddlewareProvider(tracer trace.Tracer) *MiddlewareProvider {
	return &MiddlewareProvider{tracer: tracer}
}

type MiddleWareFunc func(http.Handler) http.Handler

func (MiddlewareProvider) CompileHandlers(base http.Handler, middlewares ...MiddleWareFunc) http.Handler {
	final := base

	for i := len(middlewares) - 1; i >= 0; i-- {
		final = middlewares[i](final)
	}

	return final
}
