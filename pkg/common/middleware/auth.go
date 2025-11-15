package middleware

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/config"
	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
)

type authHeader struct {
	Token string `header:"Authorization"`
}

func AuthMiddleWave(ctx *gin.Context) {
	accessToken, err := retrieveAccessToken(ctx)
	if err != nil {
		ctx.JSON(401, gin.H{})
		ctx.Abort()
		return
	}

	payload, err := util.ValidateToken(accessToken, config.Instance.JwtAccessSecretKey, models.TokenVersion)
	if err != nil {
		ctx.JSON(401, gin.H{})
		ctx.Abort()
		return
	}

	var user models.TokenPayload
	jsonData, _ := json.Marshal(payload)
	json.Unmarshal(jsonData, &user)
	ctx.Set("user", user)

	ctx.Next()
}

func retrieveAccessToken(ctx *gin.Context) (string, error) {
	h := authHeader{}

	if err := ctx.ShouldBindHeader(&h); err != nil {
		return "", err
	}

	parts := strings.Split(h.Token, `Bearer `)

	if len(parts) != 2 {
		return "", errors.New("invalid token")
	}

	return parts[1], nil
}
