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

	routes := r.Group("/analytics/users")

	routes.POST("/", controller.getUserAnalytics)
	routes.POST("/perks", controller.getPerksAnalytics)
	routes.POST("/perks/playtime",
		middleware.OptionalAuthMiddleWave, controller.getPlaytimeHist)
	routes.POST("/perks/accuracy",
		middleware.OptionalAuthMiddleWave, controller.getAccuracyHist)
	routes.POST("/teammates",
		middleware.OptionalAuthMiddleWave, controller.getTeammates)
	routes.POST("/maps", controller.getPlayedMaps)
	routes.POST("/difficulty", controller.getDifficultyHist)
	routes.POST("/sessions",
		middleware.OptionalAuthMiddleWave, controller.getUserSessions)
	routes.POST("/lastseen",
		middleware.AuthMiddleWave, controller.getLastSeenUsers)
	routes.POST("/lastgameswithuser",
		middleware.AuthMiddleWave, controller.getLastGamesWithUser)
}
