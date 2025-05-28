package migrations

import "database/sql"

func migration_2025_05_23_0001_add_fields(db *sql.DB) {
	name := "migration_2025_05_23_0001_add_fields"

	if isMigrationExists(db, name) {
		return
	}

	_, err := db.Exec(`
 		DROP PROCEDURE IF EXISTS migration_2025_05_23_0001_add_fields;
 		CREATE PROCEDURE migration_2025_05_23_0001_add_fields()
 		BEGIN
			ALTER TABLE session 
			ADD COLUMN calc_diff REAL NOT NULL DEFAULT 0 AFTER diff;

			ALTER TABLE session_aggregated
			ADD COLUMN buffs_active_length REAL NOT NULL DEFAULT 0 AFTER zedtime_length,
			ADD COLUMN buffs_total_length REAL NOT NULL DEFAULT 0 AFTER buffs_active_length;
 		END;
 
 		CALL migration_2025_05_23_0001_add_fields();
 		DROP PROCEDURE IF EXISTS migration_2025_05_23_0001_add_fields;
 		`,
	)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
