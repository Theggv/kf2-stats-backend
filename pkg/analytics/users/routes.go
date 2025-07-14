package users

import (
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.RouterGroup,
	service *UserAnalyticsService,
	memoryStore *persist.MemoryStore,
) {
	controller := controller{
		service: service,
	}

	routes := r.Group("/analytics/")

	routes.POST("/users", controller.getUserAnalytics)
	routes.POST("/users/perks", controller.getPerksAnalytics)
	routes.POST("/users/perks/playtime", controller.getPlaytimeHist)
	routes.POST("/users/perks/accuracy", controller.getAccuracyHist)
	routes.POST("/users/teammates", controller.getTeammates)
	routes.POST("/users/maps", controller.getPlayedMaps)
}
