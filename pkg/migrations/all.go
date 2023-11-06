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
		SELECT count(*) FROM migrations WHERE name = $1`, name,
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
		INSERT INTO migrations (name) VALUES ($1)`, name,
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf("%v is performed.\n", name)
}

func ExecuteAll(db *sql.DB) {
	initTables(db)

	// migration_24_10_2023_aggr_kills(db)
	// migration_25_10_2023_alter_fkey(db)
	// migration_25_10_2023_delete_negative_shots(db)
	// migration_27_10_2023_drop_users_name_history(db)
	// migration_27_10_2023_fill_users_activity(db)
	// migration_27_10_2023_fix_game_data_cd(db)
}
