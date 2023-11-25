package migrations

import "database/sql"

func migration_2023_11_19_session_aggr_kills(db *sql.DB) {
	name := "migration_2023_11_19_session_aggr_kills"

	if isMigrationExists(db, name) {
		return
	}

	_, err := db.Exec(`
		INSERT INTO session_aggregated_kills (id, trash, medium, large, total)
		(
			SELECT
				aggr.id,
				trash,
				medium,
				large,
				total
			FROM session_aggregated aggr
			INNER JOIN (
				SELECT 
					session.id as session_id, 
					wsp.player_id as user_id, 
					wsp.perk as perk, 
					sum(kills.trash) as trash, 
					sum(kills.medium) as medium, 
					sum(kills.large) as large, 
					sum(kills.total) as total
				FROM session
				INNER JOIN wave_stats ws ON ws.session_id = session.id
				INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
				INNER JOIN aggregated_kills kills ON kills.player_stats_id = wsp.id
				WHERE session.is_completed = true
				GROUP BY session.id, wsp.player_id, wsp.perk
			) t ON aggr.session_id = t.session_id AND aggr.user_id = t.user_id AND aggr.perk = t.perk
		)
		`,
	)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
