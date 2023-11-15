package matches

import (
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.RouterGroup,
	service *MatchesService,
	memoryStore *persist.MemoryStore,
) {
	controller := matchesController{
		service: service,
	}

	routes := r.Group("/matches")

	routes.GET("/:id",
		cache.CacheByRequestURI(memoryStore, 15*time.Second),
		controller.getById)
	routes.GET("/:id/live",
		controller.getMatchLiveData)
	routes.GET("/:id/waves",
		cache.CacheByRequestURI(memoryStore, 15*time.Second),
		controller.getMatchWaves)
	routes.GET("/:id/user/:userId/stats",
		cache.CacheByRequestURI(memoryStore, 15*time.Second),
		controller.getMatchPlayerStats)
	routes.GET("/:id/summary",
		cache.CacheByRequestURI(memoryStore, 15*time.Second),
		controller.getMatchAggregatedStats)
	routes.GET("/wave/:id/stats",
		cache.CacheByRequestURI(memoryStore, 5*time.Minute),
		controller.getWavePlayersStats)
	routes.GET("/server/:id",
		cache.CacheByRequestURI(memoryStore, 15*time.Second),
		controller.getLastServerMatch)
	routes.POST("/filter", controller.filter)
}
