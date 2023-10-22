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
	routes.GET("/:id/stats", controller.getMatchStats)
	routes.GET("/:id/user/:userId/stats", controller.getMatchPlayerStats)
	routes.GET("/wave/:id/stats", controller.getWavePlayersStats)
	routes.GET("/server/:id", controller.getLastServerMatch)
	routes.POST("/filter", controller.filter)
}
