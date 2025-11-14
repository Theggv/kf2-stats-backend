package migrations

import (
	"database/sql"
	"fmt"
)

func migration_2025_11_14_0001_clean_procs(db *sql.DB) {
	name := "migration_2025_11_14_0001_clean_procs"

	if isMigrationExists(db, name) {
		return
	}

	fmt.Printf("performing %v...\n", name)

	_, err := db.Exec(`
 		DROP PROCEDURE IF EXISTS abort_old_sessions;
		DROP PROCEDURE IF EXISTS delete_empty_sessions;
		DROP PROCEDURE IF EXISTS fix_dropped_sessions;
 		`,
	)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
