package migrations

import "database/sql"

func migration_25_10_2023_delete_negative_shots(db *sql.DB) {
	name := "migration_25_10_2023_delete_negative_shots"

	if isMigrationExists(db, name) {
		return
	}

	_, err := db.Exec(`
		DELETE FROM wave_stats_player
		WHERE shots_fired < 0
		`,
	)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
