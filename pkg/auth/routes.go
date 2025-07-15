package auth

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, service *AuthService) {
	controller := authController{
		service: service,
	}

	routes := r.Group("/auth")

	routes.POST("/login", controller.login)
	routes.POST("/refresh", controller.refresh)
	routes.POST("/logout", controller.logout)
}
