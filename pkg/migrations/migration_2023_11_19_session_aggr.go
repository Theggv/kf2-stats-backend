package migrations

import "database/sql"

func migration_2023_11_19_session_aggr(db *sql.DB) {
	name := "migration_2023_11_19_session_aggr"

	if isMigrationExists(db, name) {
		return
	}

	_, err := db.Exec(`
		INSERT INTO session_aggregated (
			session_id, user_id, perk, 
			playtime_seconds, waves_played, deaths, 
			shots_fired, shots_hit, shots_hs, 
			dosh_earned, heals_given, heals_recv, 
			damage_dealt, damage_taken, 
			zedtime_count, zedtime_length)
		(
			SELECT 
				session.id as session_id, 
				wsp.player_id as user_id, 
				wsp.perk as perk, 
				sum(timestampdiff(SECOND, ws.started_at, ws.completed_at)) as playtime_seconds,
				count(*) as waves_played,
				sum(is_dead = 1) as deaths, 
				sum(shots_fired) as shots_fired, 
				sum(shots_hit) as shots_hit, 
				sum(shots_hs) as shots_hs, 
				sum(dosh_earned) as dosh_earned, 
				sum(heals_given) as heals_given, 
				sum(heals_recv) as heals_recv, 
				sum(damage_dealt) as damage_dealt, 
				sum(damage_taken) as damage_taken, 
				sum(zedtime_count) as zedtime_count, 
				sum(zedtime_length) as zedtime_length
			FROM session
			INNER JOIN wave_stats ws ON ws.session_id = session.id
			INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
			WHERE session.is_completed = true
			GROUP BY session.id, wsp.player_id, wsp.perk
		)
		`,
	)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
