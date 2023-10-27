package migrations

import "database/sql"

func migration_27_10_2023_fill_users_activity(db *sql.DB) {
	name := "migration_27_10_2023_fill_users_activity"

	if isMigrationExists(db, name) {
		return
	}

	_, err := db.Exec(`
		INSERT INTO users_activity (user_id, current_session_id, last_session_id)
		SELECT id, NULL, NULL from users`,
	)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
