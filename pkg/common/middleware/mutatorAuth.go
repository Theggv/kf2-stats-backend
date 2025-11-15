package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/config"
)

func MutatorAuthMiddleWave(ctx *gin.Context) {
	// Ignore middleware if secretToken is not set
	secretToken := config.Instance.Token

	if secretToken == "" {
		ctx.Next()
		return
	}

	accessToken, err := retrieveAccessToken(ctx)
	if err != nil {
		ctx.JSON(401, gin.H{"message": "Invalid bearer token"})
		ctx.Abort()
		return
	}

	if secretToken != accessToken {
		ctx.JSON(401, gin.H{"message": "Invalid token"})
		ctx.Abort()
		return
	}

	ctx.Next()
}
