package filter

import (
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.RouterGroup,
	service *MatchesFilterService,
	memoryStore *persist.MemoryStore,
) {
	controller := controller{
		service: service,
	}

	routes := r.Group("/matches")

	routes.POST("/filter", controller.filter)
}
