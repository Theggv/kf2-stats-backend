package util

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

func ValidateToken(tokenString string, key string) (any, error) {
	token, err := jwt.Parse(
		tokenString,
		func(t *jwt.Token) (any, error) {
			return []byte(key), nil
		},
		jwt.WithExpirationRequired(), jwt.WithIssuedAt(),
	)

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if payload, ok := claims["payload"]; ok {
			return payload, nil
		}
	}

	return nil, errors.New("failed to parse payload")
}
