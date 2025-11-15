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
	controller := matchesController{
		service: service,
	}

	routes := r.Group("/matches")

	routes.POST("/filter/new", controller.filter)
}
