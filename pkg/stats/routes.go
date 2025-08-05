package stats

import (
	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, service *StatsService) {
	controller := statsController{
		service: service,
	}

	routes := r.Group("/stats")

	routes.POST("/wave", middleware.MutatorAuthMiddleWave, controller.createWaveStats)
}
