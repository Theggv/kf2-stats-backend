package main

import (
	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/database"
	"github.com/theggv/kf2-stats-backend/pkg/common/store"
	"github.com/theggv/kf2-stats-backend/pkg/maps"
	"github.com/theggv/kf2-stats-backend/pkg/server"
	"github.com/theggv/kf2-stats-backend/pkg/session"
)

func main() {
	db := database.NewSQLiteDB()

	rootStore := store.New(db)

	r := gin.Default()
	api := r.Group("/api")

	server.RegisterRoutes(api, rootStore.Servers)
	maps.RegisterRoutes(api, rootStore.Maps)
	session.RegisterRoutes(api, rootStore.Sessions)

	r.Run("localhost:3000")
}
