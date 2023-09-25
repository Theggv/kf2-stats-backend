package server

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, serverService *ServerService) {
	controller := serverController{
		service: serverService,
	}

	routes := r.Group("/servers")

	routes.POST("/", controller.add)
	routes.GET("/", controller.getByPattern)
	routes.GET("/:id", controller.getById)
	routes.PUT("/name", controller.updateName)
}
