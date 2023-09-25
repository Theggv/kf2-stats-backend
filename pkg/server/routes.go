package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, db *sql.DB) {
	serverController := newServerController(db)

	routes := r.Group("/servers")

	routes.POST("/", serverController.add)
	routes.GET("/", serverController.getByPattern)
	routes.GET("/:id", serverController.getById)
	routes.PUT("/name", serverController.updateName)
}
