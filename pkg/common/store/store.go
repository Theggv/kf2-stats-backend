package store

import (
	"database/sql"

	"github.com/theggv/kf2-stats-backend/pkg/maps"
	"github.com/theggv/kf2-stats-backend/pkg/server"
)

type Store struct {
	Servers *server.ServerService
	Maps    *maps.MapsService
}

func New(db *sql.DB) *Store {
	return &Store{
		Servers: server.NewServerService(db),
		Maps:    maps.NewMapsService(db),
	}
}
