package maps

import (
	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, mapsService *MapsService) {
	controller := mapsController{
		service: mapsService,
	}

	routes := r.Group("/maps")

	routes.GET("/", controller.getByPattern)
	routes.GET("/:id", controller.getById)
	routes.PUT("/preview", middleware.AuthMiddleware, controller.updatePreview)
}
