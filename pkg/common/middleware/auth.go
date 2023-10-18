package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/config"
)

type authHeader struct {
	Token string `header:"Authorization"`
}

func AuthMiddleware(ctx *gin.Context) {
	// Ignore middleware if secretToken is not set
	secretToken := config.Instance.Token

	if secretToken == "" {
		ctx.Next()
		return
	}

	h := authHeader{}

	if err := ctx.ShouldBindHeader(&h); err != nil {
		ctx.JSON(401, gin.H{"message": "No auth header"})
		ctx.Abort()
		return
	}

	parts := strings.Split(h.Token, `Bearer `)

	if len(parts) < 2 {
		ctx.JSON(401, gin.H{"message": "Invalid bearer token"})
		ctx.Abort()
		return
	}

	if secretToken != parts[1] {
		ctx.JSON(401, gin.H{"message": "Invalid token"})
		ctx.Abort()
		return
	}

	ctx.Next()
}
