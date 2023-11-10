package migrations

import "database/sql"

func migration_2023_10_11_is_completed(db *sql.DB) {
	name := "migration_2023_10_11_is_completed"

	if isMigrationExists(db, name) {
		return
	}

	_, err := db.Exec(`
		ALTER TABLE session DROP COLUMN is_completed;
		ALTER TABLE session ADD COLUMN is_completed BOOLEAN GENERATED ALWAYS AS (status IN (-1,2,3,4)) STORED;
		`,
	)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
