package session

import (
	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, service *SessionService) {
	controller := sessionController{
		service: service,
	}

	routes := r.Group("/sessions")

	routes.POST("/", middleware.MutatorAuthMiddleWave, controller.create)
	routes.PUT("/status", middleware.MutatorAuthMiddleWave, controller.updateStatus)
	routes.PUT("/game-data", middleware.MutatorAuthMiddleWave, controller.updateGameData)
	routes.GET("/demo/:id", controller.getDemo)
	routes.POST("/demo", middleware.MutatorAuthMiddleWave, controller.uploadDemo)
}
