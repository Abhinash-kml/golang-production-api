package util

import (
	"fmt"
	"time"

	"github.com/abhinash-kml/go-api-server/config"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func CreateJwtToken(secret, issuer, subject string, audience []string, expiry time.Duration) (string, error) {
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

func VerifyJwtToken(c *config.AuthTokenConfig, tokenString string, tokenType string) (*jwt.Token, *jwt.RegisteredClaims, error) {
	var tokenConfig *config.TokenConfig
	switch tokenType {
	case "access":
		{
			tokenConfig = &c.AccessToken
		}
	case "refresh":
		{
			tokenConfig = &c.RefreshToken
		}
	}

	tempClaims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, tempClaims, func(t *jwt.Token) (any, error) {
		return []byte(tokenConfig.Secret), nil
	},
		jwt.WithIssuer(tokenConfig.Issuer),
		jwt.WithExpirationRequired(),
		jwt.WithNotBeforeRequired(),
		jwt.WithAllAudiences(tokenConfig.Audience))

	if err != nil {
		zap.L().Info("JWT Parse Error", zap.Error(err))
		return nil, nil, err
	}

	if !token.Valid {
		return nil, nil, fmt.Errorf("Token invalid")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, nil, fmt.Errorf("Token type assertion failed")
	}

	return token, claims, nil
}
