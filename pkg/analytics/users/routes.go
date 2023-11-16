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
}
