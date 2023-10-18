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

	routes.POST("/", middleware.AuthMiddleware, controller.create)
	routes.GET("/:id", controller.getById)
	routes.POST("/filter", controller.filter)
	routes.PUT("/status", middleware.AuthMiddleware, controller.updateStatus)
	routes.PUT("/game-data", middleware.AuthMiddleware, controller.updateGameData)
}
