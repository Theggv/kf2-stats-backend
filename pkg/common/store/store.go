package store

import (
	"database/sql"

	analyticsMaps "github.com/theggv/kf2-stats-backend/pkg/analytics/maps"
	analyticsPerks "github.com/theggv/kf2-stats-backend/pkg/analytics/perks"
	analyticsServer "github.com/theggv/kf2-stats-backend/pkg/analytics/server"
	analyticsUsers "github.com/theggv/kf2-stats-backend/pkg/analytics/users"
	"github.com/theggv/kf2-stats-backend/pkg/auth"
	"github.com/theggv/kf2-stats-backend/pkg/common/config"
	"github.com/theggv/kf2-stats-backend/pkg/common/steamapi"
	"github.com/theggv/kf2-stats-backend/pkg/leaderboards"
	"github.com/theggv/kf2-stats-backend/pkg/maps"
	"github.com/theggv/kf2-stats-backend/pkg/matches"
	"github.com/theggv/kf2-stats-backend/pkg/server"
	"github.com/theggv/kf2-stats-backend/pkg/session"
	"github.com/theggv/kf2-stats-backend/pkg/stats"
	"github.com/theggv/kf2-stats-backend/pkg/users"
)

type Store struct {
	Db *sql.DB

	Auth     *auth.AuthService
	Servers  *server.ServerService
	Maps     *maps.MapsService
	Sessions *session.SessionService
	Stats    *stats.StatsService
	Users    *users.UserService
	Matches  *matches.MatchesService
	SteamApi *steamapi.SteamApiUserService

	AnalyticsMaps   *analyticsMaps.MapAnalyticsService
	AnalyticsServer *analyticsServer.ServerAnalyticsService
	AnalyticsPerks  *analyticsPerks.PerksAnalyticsService
	AnalyticsUsers  *analyticsUsers.UserAnalyticsService

	LeaderBoards *leaderboards.LeaderBoardsService
}

func New(db *sql.DB, config *config.AppConfig) *Store {
	store := Store{
		Db: db,

		Auth:     auth.NewAuthService(db),
		Servers:  server.NewServerService(db),
		Maps:     maps.NewMapsService(db),
		Sessions: session.NewSessionService(db),
		Stats:    stats.NewStatsService(db),
		Users:    users.NewUserService(db),
		Matches:  matches.NewMatchesService(db),
		SteamApi: steamapi.NewSteamApiUserService(config.SteamApiKey),

		AnalyticsMaps:   analyticsMaps.NewMapAnalyticsService(db),
		AnalyticsServer: analyticsServer.NewServerAnalyticsService(db),
		AnalyticsPerks:  analyticsPerks.NewPerksAnalyticsService(db),
		AnalyticsUsers:  analyticsUsers.NewUserAnalyticsService(db),

		LeaderBoards: leaderboards.NewLeaderBoardsService(db),
	}

	store.Auth.Inject(store.Users, store.SteamApi)
	store.Servers.Inject(store.Users)
	store.Stats.Inject(store.Users)
	store.Sessions.Inject(store.Maps, store.Servers, store.Users)
	store.Matches.Inject(store.Users, store.Sessions, store.Maps, store.Servers, store.SteamApi)
	store.Users.Inject(store.SteamApi)
	store.AnalyticsUsers.Inject(store.Users)
	store.LeaderBoards.Inject(store.Users)

	return &store
}
