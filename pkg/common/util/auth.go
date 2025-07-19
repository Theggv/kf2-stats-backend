package util

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

func SignToken(payload any, key string, expiresIn string, version int) (string, error) {
	duration, err := time.ParseDuration(expiresIn)
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"payload": payload,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(duration).Unix(),
	}

	if version > 0 {
		claims["version"] = version
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(key))
}

// use version=0 to skip version check
func ValidateToken(tokenString string, key string, version int) (any, error) {
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
		if version > 0 {
			version, ok := claims["version"]
			if !ok {
				return nil, errors.New("token version mismatch")
			}

			if version.(float64) != models.TokenVersion {
				return nil, errors.New("token version mismatch")
			}
		}

		if payload, ok := claims["payload"]; ok {
			return payload, nil
		}
	}

	return nil, errors.New("failed to parse payload")
}

func SetCookies(ctx *gin.Context, refreshToken string, expiresIn string) error {
	duration, err := time.ParseDuration(expiresIn)
	if err != nil {
		return err
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Path:     "/",
		Domain:   "localhost",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(duration.Seconds()),
		Secure:   false,
		HttpOnly: true,
	})

	return nil
}

func GetUserFromCtx(ctx *gin.Context) (*models.TokenPayload, bool) {
	if payload, ok := ctx.Get("user"); ok {
		user := payload.(models.TokenPayload)
		return &user, true
	}

	return nil, false
}
