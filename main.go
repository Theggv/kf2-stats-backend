package main

import (
	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/database"
	"github.com/theggv/kf2-stats-backend/pkg/common/store"
	"github.com/theggv/kf2-stats-backend/pkg/server"
)

func main() {
	db := database.NewSQLiteDB()

	rootStore := store.New(db)

	r := gin.Default()
	api := r.Group("/api")

	server.RegisterRoutes(api, rootStore.Servers)

	r.Run("localhost:3000")
}
