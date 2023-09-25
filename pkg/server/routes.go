package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, db *sql.DB) {
	serverController := NewServerController(db)

	routes := r.Group("/servers")

	routes.POST("/", serverController.Add)
	routes.GET("/", serverController.GetByPattern)
	routes.GET("/:id", serverController.GetById)
}
