package main

import (
	"time"

	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/config"
	"github.com/theggv/kf2-stats-backend/pkg/common/database/mysql"
	"github.com/theggv/kf2-stats-backend/pkg/common/store"
	"github.com/theggv/kf2-stats-backend/pkg/migrations"
	"github.com/theggv/kf2-stats-backend/pkg/router"

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
	memoryStore := persist.NewMemoryStore(5 * time.Minute)

	// Run migrations
	migrations.ExecuteAll(db)

	r := gin.Default()

	// Setup cors
	r.Use(cors.Default())

	// Register api routes
	router.RegisterApiRoutes(r, rootStore, memoryStore)

	// Setup swagger
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Run app
	r.Run(config.ServerAddr)
}
