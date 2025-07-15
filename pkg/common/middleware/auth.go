package middleware

import (
	"encoding/json"
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
	h := authHeader{}

	if err := ctx.ShouldBindHeader(&h); err != nil {
		ctx.JSON(401, gin.H{})
		ctx.Abort()
		return
	}

	parts := strings.Split(h.Token, `Bearer `)

	if len(parts) < 2 {
		ctx.JSON(401, gin.H{})
		ctx.Abort()
		return
	}

	accessToken := parts[1]
	payload, err := util.ValidateToken(accessToken, config.Instance.JwtAccessSecretKey)
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
