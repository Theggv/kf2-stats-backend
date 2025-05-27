package migrations

import (
	"database/sql"
	"fmt"
)

func migration_2025_05_27_0001_migrate_leaderboard(db *sql.DB) {
	name := "migration_2025_05_27_0001_migrate_leaderboard"

	if isMigrationExists(db, name) {
		return
	}

	fmt.Printf("performing %v...\n", name)

	_, err := db.Exec(`
 		DROP PROCEDURE IF EXISTS migration_2025_05_27_0001_migrate_leaderboard;
 		CREATE PROCEDURE migration_2025_05_27_0001_migrate_leaderboard()
 		BEGIN
			CALL fill_weekly_user_stats();
 		END;
 
 		CALL migration_2025_05_27_0001_migrate_leaderboard();
 		DROP PROCEDURE IF EXISTS migration_2025_05_27_0001_migrate_leaderboard;
 		`,
	)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
