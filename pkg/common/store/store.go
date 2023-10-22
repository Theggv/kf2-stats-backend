package store

import (
	"database/sql"

	"github.com/theggv/kf2-stats-backend/pkg/common/config"
	"github.com/theggv/kf2-stats-backend/pkg/common/steamapi"
	"github.com/theggv/kf2-stats-backend/pkg/maps"
	"github.com/theggv/kf2-stats-backend/pkg/matches"
	"github.com/theggv/kf2-stats-backend/pkg/server"
	"github.com/theggv/kf2-stats-backend/pkg/session"
	"github.com/theggv/kf2-stats-backend/pkg/stats"
	"github.com/theggv/kf2-stats-backend/pkg/users"
	"github.com/theggv/kf2-stats-backend/pkg/views"
)

type Store struct {
	Servers  *server.ServerService
	Maps     *maps.MapsService
	Sessions *session.SessionService
	Stats    *stats.StatsService
	Users    *users.UserService
	Matches  *matches.MatchesService
	Views    *views.ViewsService
	SteamApi *steamapi.SteamApiUserService
}

func New(db *sql.DB, config *config.AppConfig) *Store {
	store := Store{
		Servers:  server.NewServerService(db),
		Maps:     maps.NewMapsService(db),
		Sessions: session.NewSessionService(db),
		Stats:    stats.NewStatsService(db),
		Users:    users.NewUserService(db),
		Matches:  matches.NewMatchesService(db),
		Views:    views.NewViewsService(db),
		SteamApi: steamapi.NewSteamApiUserService(config.SteamApiKey),
	}

	store.Stats.Inject(store.Users)
	store.Matches.Inject(store.Sessions, store.Maps, store.Servers, store.SteamApi)

	return &store
}
