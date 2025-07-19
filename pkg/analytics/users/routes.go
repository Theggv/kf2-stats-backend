package users

import (
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/middleware"
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
	routes.POST("/users/perks/playtime",
		middleware.OptionalAuthMiddleWave, controller.getPlaytimeHist)
	routes.POST("/users/perks/accuracy",
		middleware.OptionalAuthMiddleWave, controller.getAccuracyHist)
	routes.POST("/users/teammates",
		middleware.OptionalAuthMiddleWave, controller.getTeammates)
	routes.POST("/users/maps", controller.getPlayedMaps)
	routes.POST("/users/lastseen",
		middleware.AuthMiddleWave, controller.getLastSeenUsers)
	routes.POST("/users/lastgameswithuser",
		middleware.AuthMiddleWave, controller.getLastGamesWithUser)
}
