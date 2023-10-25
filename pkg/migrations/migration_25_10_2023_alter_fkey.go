package migrations

import "database/sql"

func migration_25_10_2023_aggr_kills(db *sql.DB) {
	name := "migration_25_10_2023_aggr_kills"

	if isMigrationExists(db, name) {
		return
	}

	_, err := db.Exec(`
		pragma writable_schema=1;
		update SQLITE_MASTER set sql = replace(sql, 'REFERENCES user(id)',
			'REFERENCES users(id)'
		) where name = 'wave_stats_player' and type = 'table';
		pragma writable_schema=0;
		`,
	)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
