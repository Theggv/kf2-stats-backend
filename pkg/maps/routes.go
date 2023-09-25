package maps

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, mapsService *MapsService) {
	controller := mapsController{
		service: mapsService,
	}

	routes := r.Group("/maps")

	routes.POST("/", controller.add)
	routes.GET("/", controller.getByPattern)
	routes.GET("/:id", controller.getById)
	routes.PUT("/name", controller.updatePreview)
}
