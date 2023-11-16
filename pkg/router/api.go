package router

import (
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	analyticsMaps "github.com/theggv/kf2-stats-backend/pkg/analytics/maps"
	analyticsPerks "github.com/theggv/kf2-stats-backend/pkg/analytics/perks"
	analyticsServer "github.com/theggv/kf2-stats-backend/pkg/analytics/server"
	analyticsUsers "github.com/theggv/kf2-stats-backend/pkg/analytics/users"
	"github.com/theggv/kf2-stats-backend/pkg/common/store"
	"github.com/theggv/kf2-stats-backend/pkg/maps"
	"github.com/theggv/kf2-stats-backend/pkg/matches"
	"github.com/theggv/kf2-stats-backend/pkg/server"
	"github.com/theggv/kf2-stats-backend/pkg/session"
	"github.com/theggv/kf2-stats-backend/pkg/stats"
	"github.com/theggv/kf2-stats-backend/pkg/users"
)

func RegisterApiRoutes(r *gin.Engine, store *store.Store, memoryStore *persist.MemoryStore) {
	api := r.Group("/api")

	server.RegisterRoutes(api, store.Servers)
	maps.RegisterRoutes(api, store.Maps)
	session.RegisterRoutes(api, store.Sessions)
	stats.RegisterRoutes(api, store.Stats)
	users.RegisterRoutes(api, store.Users)
	matches.RegisterRoutes(api, store.Matches, memoryStore)

	analyticsMaps.RegisterRoutes(api, store.AnalyticsMaps, memoryStore)
	analyticsServer.RegisterRoutes(api, store.AnalyticsServer, memoryStore)
	analyticsPerks.RegisterRoutes(api, store.AnalyticsPerks, memoryStore)
	analyticsUsers.RegisterRoutes(api, store.AnalyticsUsers, memoryStore)
}
