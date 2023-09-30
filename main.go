package main

import (
	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/database"
	"github.com/theggv/kf2-stats-backend/pkg/common/store"
	"github.com/theggv/kf2-stats-backend/pkg/maps"
	"github.com/theggv/kf2-stats-backend/pkg/server"
	"github.com/theggv/kf2-stats-backend/pkg/session"
	"github.com/theggv/kf2-stats-backend/pkg/stats"
	"github.com/theggv/kf2-stats-backend/pkg/users"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/theggv/kf2-stats-backend/docs"
)

// @title KF2 Stats Backend API
// @version 1.0

// @BasePath /api
func main() {
	db := database.NewSQLiteDB()

	rootStore := store.New(db)

	r := gin.Default()
	api := r.Group("/api")

	server.RegisterRoutes(api, rootStore.Servers)
	maps.RegisterRoutes(api, rootStore.Maps)
	session.RegisterRoutes(api, rootStore.Sessions)
	stats.RegisterRoutes(api, rootStore.Stats)
	users.RegisterRoutes(api, rootStore.Users)

	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run("localhost:3000")
}
