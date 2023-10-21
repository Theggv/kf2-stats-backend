package views

import "database/sql"

type ViewsService struct {
	db *sql.DB
}

func (s *ViewsService) initTables() {
	s.db.Exec(`
	CREATE VIEW view_indexes as
	SELECT
		session.id as session_id,
		wave_stats.id as wave_stats_id,
		wave_stats_player.id as wave_stats_player_id,
		session.server_id as server_id,
		session.map_id as map_id,
		wave_stats_player.player_id as player_id
	FROM session
	INNER JOIN wave_stats ON wave_stats.session_id = session.id
	INNER JOIN wave_stats_player ON wave_stats_player.stats_id = wave_stats.id
	`)
}

func NewViewsService(db *sql.DB) *ViewsService {
	service := ViewsService{
		db: db,
	}

	service.initTables()

	return &service
}

func (s *ViewsService) query() {

}
