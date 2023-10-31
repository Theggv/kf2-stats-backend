package maps

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, service *MapAnalyticsService) {
	controller := controller{
		service: service,
	}

	routes := r.Group("/analytics/")

	routes.POST("/maps", controller.getMapAnalytics)
}
