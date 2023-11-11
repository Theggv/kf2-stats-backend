package main

import (
	"time"

	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	analyticsMaps "github.com/theggv/kf2-stats-backend/pkg/analytics/maps"
	analyticsPerks "github.com/theggv/kf2-stats-backend/pkg/analytics/perks"
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
)

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
	api := r.Group("/api")

	server.RegisterRoutes(api, rootStore.Servers)
	maps.RegisterRoutes(api, rootStore.Maps)
	session.RegisterRoutes(api, rootStore.Sessions)
	stats.RegisterRoutes(api, rootStore.Stats)
	users.RegisterRoutes(api, rootStore.Users)
	matches.RegisterRoutes(api, rootStore.Matches, memoryStore)

	analyticsMaps.RegisterRoutes(api, rootStore.AnalyticsMaps, memoryStore)
	analyticsServer.RegisterRoutes(api, rootStore.AnalyticsServer, memoryStore)
	analyticsPerks.RegisterRoutes(api, rootStore.AnalyticsPerks, memoryStore)

	// Run app
	r.Run(config.ServerAddr)
}
