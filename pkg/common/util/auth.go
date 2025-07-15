package util

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

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

func GetUserFromCtx(ctx *gin.Context) *models.TokenPayload {
	payload, _ := ctx.Get("user")
	user := payload.(models.TokenPayload)
	return &user
}
