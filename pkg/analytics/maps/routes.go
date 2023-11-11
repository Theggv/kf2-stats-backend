package maps

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
	service *MapAnalyticsService,
	memoryStore *persist.MemoryStore,
) {
	controller := controller{
		service: service,
	}

	routes := r.Group("/analytics/")

	routes.POST("/maps",
		cache.Cache(memoryStore, 5*time.Minute,
			strategy.CacheByRequestBody[MapAnalyticsRequest](func(req MapAnalyticsRequest) string {
				return fmt.Sprintf("%v/%v/%v/%v",
					req.ServerId, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"), req.Limit)
			}),
		),
		controller.getMapAnalytics)
}
