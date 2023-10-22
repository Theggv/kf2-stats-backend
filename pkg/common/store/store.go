package store

import (
	"database/sql"

	"github.com/theggv/kf2-stats-backend/pkg/common/config"
	"github.com/theggv/kf2-stats-backend/pkg/common/steamapi"
	"github.com/theggv/kf2-stats-backend/pkg/maps"
	"github.com/theggv/kf2-stats-backend/pkg/server"
	"github.com/theggv/kf2-stats-backend/pkg/session"
	"github.com/theggv/kf2-stats-backend/pkg/stats"
	"github.com/theggv/kf2-stats-backend/pkg/users"
)

type Store struct {
	Servers  *server.ServerService
	Maps     *maps.MapsService
	Sessions *session.SessionService
	Stats    *stats.StatsService
	Users    *users.UserService
	SteamApi *steamapi.SteamApiUserService
}

func New(db *sql.DB, config *config.AppConfig) *Store {
	store := Store{
		Servers:  server.NewServerService(db),
		Maps:     maps.NewMapsService(db),
		Sessions: session.NewSessionService(db),
		Stats:    stats.NewStatsService(db),
		Users:    users.NewUserService(db),
		SteamApi: steamapi.NewSteamApiUserService(config.SteamApiKey),
	}

	store.Stats.Inject(store.Users)
	store.Sessions.Inject(store.Maps, store.Servers)

	return &store
}
