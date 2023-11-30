package server

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
	service *ServerAnalyticsService,
	memoryStore *persist.MemoryStore,
) {
	controller := controller{
		service: service,
	}

	routes := r.Group("/analytics/")

	routes.POST("/server/session/count",
		cache.Cache(memoryStore, 5*time.Minute,
			strategy.CacheByRequestBody[SessionCountRequest](func(req SessionCountRequest) string {
				return fmt.Sprintf("%v/%v/%v/%v",
					req.ServerId, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"), req.Period)
			}),
		),
		controller.getSessionCount)
	routes.POST("/server/usage",
		cache.Cache(memoryStore, 5*time.Minute,
			strategy.CacheByRequestBody[UsageInMinutesRequest](func(req UsageInMinutesRequest) string {
				return fmt.Sprintf("%v/%v/%v/%v",
					req.ServerId, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"), req.Period)
			}),
		),
		controller.getUsageInMinutes)
	routes.POST("/server/online",
		cache.Cache(memoryStore, 5*time.Minute,
			strategy.CacheByRequestBody[PlayersOnlineRequest](func(req PlayersOnlineRequest) string {
				return fmt.Sprintf("%v/%v/%v/%v",
					req.ServerId, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"), req.Period)
			}),
		),
		controller.getPlayersOnline)

	routes.GET("/server/popular",
		cache.CacheByRequestURI(memoryStore, 5*time.Minute),
		controller.getPopularServers)
	routes.GET("/server/current-online",
		cache.CacheByRequestURI(memoryStore, 1*time.Minute),
		controller.getCurrentOnline)
}
