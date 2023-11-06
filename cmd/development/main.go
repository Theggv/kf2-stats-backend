package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	analyticsMaps "github.com/theggv/kf2-stats-backend/pkg/analytics/maps"
	analyticsServer "github.com/theggv/kf2-stats-backend/pkg/analytics/server"
	"github.com/theggv/kf2-stats-backend/pkg/common/config"
	"github.com/theggv/kf2-stats-backend/pkg/common/database/mysql"
	"github.com/theggv/kf2-stats-backend/pkg/common/store"
	"github.com/theggv/kf2-stats-backend/pkg/maps"
	"github.com/theggv/kf2-stats-backend/pkg/matches"
	"github.com/theggv/kf2-stats-backend/pkg/migrations"
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
	config := config.Instance
	db := mysql.NewDBInstance(
		config.DBUser, config.DBPassword, config.DBHost, config.DBName, config.DBPort,
	)

	rootStore := store.New(db, config)

	// Run migrations
	migrations.ExecuteAll(db)

	r := gin.Default()

	// Setup cors
	r.Use(cors.Default())

	// Register api routes
	api := r.Group("/api")

	server.RegisterRoutes(api, rootStore.Servers)
	maps.RegisterRoutes(api, rootStore.Maps)
	session.RegisterRoutes(api, rootStore.Sessions)
	stats.RegisterRoutes(api, rootStore.Stats)
	users.RegisterRoutes(api, rootStore.Users)
	matches.RegisterRoutes(api, rootStore.Matches)

	analyticsMaps.RegisterRoutes(api, rootStore.AnalyticsMaps)
	analyticsServer.RegisterRoutes(api, rootStore.AnalyticsServer)

	// Setup swagger
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Run app
	r.Run(config.ServerAddr)
}
