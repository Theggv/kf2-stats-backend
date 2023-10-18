package server

import (
	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, serverService *ServerService) {
	controller := serverController{
		service: serverService,
	}

	routes := r.Group("/servers")

	routes.POST("/", middleware.AuthMiddleware, controller.add)
	routes.GET("/", controller.getByPattern)
	routes.GET("/:id", controller.getById)
	routes.PUT("/name", middleware.AuthMiddleware, controller.updateName)
}
