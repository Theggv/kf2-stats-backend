package middleware

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/config"
	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
)

func OptionalAuthMiddleWave(ctx *gin.Context) {
	accessToken, err := retrieveAccessToken(ctx)
	if err != nil {
		ctx.Next()
		return
	}

	payload, err := util.ValidateToken(accessToken, config.Instance.JwtAccessSecretKey, models.TokenVersion)
	if err != nil {
		ctx.Next()
		return
	}

	var user models.TokenPayload
	jsonData, _ := json.Marshal(payload)
	json.Unmarshal(jsonData, &user)
	ctx.Set("user", user)

	ctx.Next()
}
