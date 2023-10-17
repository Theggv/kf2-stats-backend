package stats

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, service *StatsService) {
	controller := statsController{
		service: service,
	}

	routes := r.Group("/stats")

	routes.POST("/wave/player", controller.createWavePlayerStats)
	routes.POST("/wave/cd", controller.createWaveStatsCD)
}
