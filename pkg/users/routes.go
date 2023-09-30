package users

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, service *UserService) {
	controller := userController{
		service: service,
	}

	routes := r.Group("/users")

	routes.POST("/", controller.create)
}
