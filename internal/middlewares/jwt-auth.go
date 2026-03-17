package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (m *MiddlewareProvider) JwtAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := m.tracer.Start(r.Context(), "middleware.JwtAuth")
		defer span.End()

		// Auth logic
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) < 2 || len(parts) > 2 || parts[0] != "Bearer" {
			http.Error(w, "Bad token format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		span.SetAttributes(attribute.String("token", tokenString))

		tempClaims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenString, tempClaims, func(t *jwt.Token) (any, error) {
			return []byte("my-special-secret"), nil
		},
			jwt.WithIssuer("my-app"),
			jwt.WithExpirationRequired(),
			jwt.WithNotBeforeRequired(),
			jwt.WithAllAudiences("www.myapp.com"))

		if err != nil {
			http.Error(w, fmt.Sprintf("%+v", err), http.StatusUnauthorized)
			fmt.Printf("JWT Error: %+v\n", err)
			span.RecordError(err)
			span.SetStatus(codes.Error, "token parsing failed")
			return
		}

		if !token.Valid {
			http.Error(w, "Token is now valid meow meow", http.StatusUnauthorized)
			span.SetAttributes(attribute.Bool("token.valid", false))
			return
		}

		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok {
			http.Error(w, "Failed to parse jwt token meow meow", http.StatusUnauthorized)
			span.SetAttributes(attribute.Bool("token.ourcustomtype", false))
			return
		}

		// Do real work here
		fmt.Println(claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
