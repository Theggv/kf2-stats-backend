package main

import (
	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/database"
	"github.com/theggv/kf2-stats-backend/pkg/server"
)

func main() {
	db := database.NewSQLiteDB()

	r := gin.Default()
	api := r.Group("/api")

	server.RegisterRoutes(api, db)

	r.Run("localhost:3000")
}
