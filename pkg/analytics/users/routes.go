package users

import (
	"fmt"
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/strategy"
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
	routes.POST("/users/top",
		cache.Cache(memoryStore, 5*time.Minute,
			strategy.CacheByRequestBody[GetUsersTopRequest](func(req GetUsersTopRequest) string {
				return fmt.Sprintf("%v/%v/%v/%v",
					req.Type, req.Perk, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
			}),
		),
		controller.getUsersTop)
}
