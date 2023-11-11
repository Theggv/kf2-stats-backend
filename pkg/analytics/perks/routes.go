package perks

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
	service *PerksAnalyticsService,
	memoryStore *persist.MemoryStore,
) {
	controller := controller{
		service: service,
	}

	routes := r.Group("/analytics/")

	routes.POST("/perks/playtime",
		cache.Cache(memoryStore, 5*time.Minute,
			strategy.CacheByRequestBody[PerksPlayTimeRequest](func(req PerksPlayTimeRequest) string {
				return fmt.Sprintf("%v/%v/%v/%v",
					req.ServerId, req.UserId, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
			}),
		),
		controller.getPerksPlayTime)
	routes.POST("/perks/kills",
		cache.Cache(memoryStore, 5*time.Minute,
			strategy.CacheByRequestBody[PerksKillsRequest](func(req PerksKillsRequest) string {
				return fmt.Sprintf("%v/%v/%v/%v",
					req.ServerId, req.UserId, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
			}),
		),
		controller.getPerksKills)
}
