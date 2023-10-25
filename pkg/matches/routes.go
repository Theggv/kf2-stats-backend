package matches

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, service *MatchesService) {
	controller := matchesController{
		service: service,
	}

	routes := r.Group("/matches")

	routes.GET("/:id", controller.getById)
	routes.GET("/:id/waves", controller.getMatchWaves)
	routes.GET("/:id/user/:userId/stats", controller.getMatchPlayerStats)
	routes.GET("/:id/summary", controller.getMatchAggregatedStats)
	routes.GET("/wave/:id/stats", controller.getWavePlayersStats)
	routes.GET("/server/:id", controller.getLastServerMatch)
	routes.POST("/filter", controller.filter)
}
