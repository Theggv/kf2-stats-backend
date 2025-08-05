package migrations

import (
	"database/sql"
	"fmt"
)

func initTables(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			name VARCHAR(128) NOT NULL PRIMARY KEY
		);`,
	)

	if err != nil {
		panic(err)
	}
}

func isMigrationExists(db *sql.DB, name string) bool {
	row := db.QueryRow(`
		SELECT count(*) FROM migrations WHERE name = ?`, name,
	)

	var count int
	err := row.Scan(&count)

	if err != nil {
		panic(err)
	}

	return count > 0
}

func writeMigration(db *sql.DB, name string) {
	_, err := db.Exec(`
		INSERT INTO migrations (name) VALUES (?)`, name,
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf("%v is performed.\n", name)
}

func ExecuteAll(db *sql.DB) {
	initTables(db)

	migration_2025_04_11_rename_tables(db)
	migration_2025_04_11_update_fields(db)
  migration_2025_05_02_idx_session_status(db)
	migration_2025_05_23_0001_add_fields(db)
	migration_2025_05_27_0001_migrate_leaderboard(db)
}
