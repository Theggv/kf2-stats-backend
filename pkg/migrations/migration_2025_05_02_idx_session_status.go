package migrations

import "database/sql"

func migration_2025_05_02_idx_session_status(db *sql.DB) {
	name := "migration_2025_05_02_idx_session_status"

	if isMigrationExists(db, name) {
		return
	}

	_, err := db.Exec(`
 		DROP PROCEDURE IF EXISTS migration_2025_05_02_idx_session_status;
 		CREATE PROCEDURE migration_2025_05_02_idx_session_status()
 		BEGIN
 			IF NOT EXISTS (
				SELECT * 
				FROM INFORMATION_SCHEMA.STATISTICS 
				WHERE TABLE_NAME = 'session' AND INDEX_NAME = 'idx_session_status'
			) THEN
			 	CREATE INDEX idx_session_status ON session(status);
 			END IF;
 		END;
 
 		CALL migration_2025_05_02_idx_session_status();
 		DROP PROCEDURE IF EXISTS migration_2025_05_02_idx_session_status;
 		`,
	)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
