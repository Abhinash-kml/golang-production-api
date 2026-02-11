package middlewares

import (
	"fmt"
	"net/http"
)

type MiddleWare func(http.Handler) http.Handler

func CompileHandlers(base http.Handler, middlewares ...MiddleWare) http.Handler {
	final := base

	for i := len(middlewares) - 1; i >= 0; i-- {
		final = middlewares[i](final)
	}

	return final
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("IP: %s Method: %s Path: %s", r.RemoteAddr, r.Method, r.URL.Path)

		next.ServeHTTP(w, r)
	})
}
