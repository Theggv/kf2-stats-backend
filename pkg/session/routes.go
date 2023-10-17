package session

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, service *SessionService) {
	controller := sessionController{
		service: service,
	}

	routes := r.Group("/sessions")

	routes.POST("/", controller.create)
	routes.GET("/:id", controller.getById)
	routes.POST("/filter", controller.filter)
	routes.PUT("/status", controller.updateStatus)
	routes.PUT("/game-data", controller.updateGameData)
}
