package store

import (
	"database/sql"

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
}

func New(db *sql.DB) *Store {
	return &Store{
		Servers:  server.NewServerService(db),
		Maps:     maps.NewMapsService(db),
		Sessions: session.NewSessionService(db),
		Stats:    stats.NewStatsService(db),
		Users:    users.NewUserService(db),
	}
}
