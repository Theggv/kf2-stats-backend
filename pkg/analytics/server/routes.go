package server

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, service *ServerAnalyticsService) {
	controller := controller{
		service: service,
	}

	routes := r.Group("/analytics/")

	routes.POST("/server/session/count", controller.getSessionCount)
	routes.POST("/server/usage", controller.getUsageInMinutes)
	routes.POST("/server/online", controller.getPlayersOnline)
}
