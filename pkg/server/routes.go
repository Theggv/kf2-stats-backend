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

	routes.GET("/", controller.getByPattern)
	routes.GET("/:id", controller.getById)
	routes.GET("/:id/last-session", controller.getLastSession)
	routes.PUT("/name", middleware.AuthMiddleware, controller.updateName)
	routes.POST("/users/recent", controller.getRecentUsers)
}
