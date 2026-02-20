package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func JwtAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Auth logic
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
		}

		tokenString := strings.Split(authHeader, " ")
		if len(tokenString) < 2 || len(tokenString) > 2 || tokenString[0] != "Bearer" {
			http.Error(w, "Bad token format", http.StatusUnauthorized)
			return
		}

		tempClaims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenString[1], tempClaims, func(t *jwt.Token) (any, error) {
			return []byte("my-special-secret"), nil
		},
			jwt.WithIssuer("my-app"),
			jwt.WithExpirationRequired(),
			jwt.WithNotBeforeRequired(),
			jwt.WithAllAudiences("www.myapp.com"))

		if err != nil {
			http.Error(w, fmt.Sprintf("%+v", err), http.StatusUnauthorized)
			fmt.Printf("JWT Error: %+v\n", err)
			return
		}

		if !token.Valid {
			http.Error(w, "Token is now valid meow meow", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok {
			http.Error(w, "Failed to parse jwt token meow meow", http.StatusUnauthorized)
			return
		}

		// Do real work here
		fmt.Println(claims)

		next.ServeHTTP(w, r)
	})
}
