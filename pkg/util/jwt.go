package util

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateToken(secret, issuer, subject string, audience []string, expiry time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   subject,
		Audience:  jwt.ClaimStrings(audience),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        "meow",
	})

	signedToken, err := token.SignedString([]byte(secret))
	return signedToken, err
}
