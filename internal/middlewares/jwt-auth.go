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
		if len(tokenString) < 2 || tokenString[0] != "Bearer" {
			http.Error(w, "Bad token format", http.StatusUnauthorized)
		}

		token, err := jwt.ParseWithClaims(tokenString[1], jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
			return []byte("my-secret-key"), nil
		})
		if err != nil {
			http.Error(w, "Failed to parse jwt token", http.StatusUnauthorized)
			fmt.Println("Failed to Parse jwt token")
		}

		claims, ok := token.Claims.(jwt.RegisteredClaims)
		if !ok {
			fmt.Println("Failed to type assert claims to custom claims type")
			http.Error(w, "Failed to parse jwt token", http.StatusUnauthorized)
		}

		// Do real work here
		fmt.Println(claims)

		next.ServeHTTP(w, r)
	})
}
