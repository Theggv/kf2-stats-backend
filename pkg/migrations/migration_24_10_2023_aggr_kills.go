package migrations

import "database/sql"

func migration_24_10_2023_aggr_kills(db *sql.DB) {
	name := "migration_24_10_2023_aggr_kills"

	if isMigrationExists(db, name) {
		return
	}

	_, err := db.Exec(`
		INSERT OR IGNORE INTO aggregated_kills (player_stats_id, trash, medium, large, total)
		SELECT 
			player_stats_id,
			cyst + alpha_clot + slasher + stalker + crawler + gorefast + rioter + elite_crawler + gorefiend,
			siren + bloat + edar + husk_n + husk_b, 
			scrake + fp + qp, 
			cyst + alpha_clot + slasher + stalker + crawler + gorefast + rioter + elite_crawler + gorefiend + 
			siren + bloat + edar + husk_n + husk_b + 
			scrake + fp + qp + boss 
		FROM wave_stats_player_kills
		`,
	)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
