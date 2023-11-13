package users

import (
	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, service *UserService) {
	controller := userController{
		service: service,
	}

	routes := r.Group("/users")

	routes.POST("/", middleware.AuthMiddleware, controller.create)
	routes.GET("/:id/detailed", controller.getUserDetailed)
	routes.POST("/filter", controller.filter)
	routes.POST("/sessions/recent", controller.getRecentSessions)
}
