package migrations

import "database/sql"

func migration_27_10_2023_drop_users_name_history(db *sql.DB) {
	name := "migration_27_10_2023_drop_users_name_history"

	if isMigrationExists(db, name) {
		return
	}

	_, err := db.Exec(`DROP TABLE IF EXISTS users_name_history;`)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
