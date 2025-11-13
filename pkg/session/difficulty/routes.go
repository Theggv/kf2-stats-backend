package difficulty

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, service *DifficultyCalculatorService) {
	controller := controller{
		service: service,
	}

	routes := r.Group("/sessions/difficulty")

	routes.GET("/:id", controller.getById)
	routes.POST("/:id", controller.addToQueue)
	routes.GET("/:id/check", controller.checkIfQueued)

	routes.POST("/server", controller.recalculateAll)
	routes.POST("/server/:id", controller.recalculateByServerId)
}
