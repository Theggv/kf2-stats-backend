package migrations

import "database/sql"

func migration_2023_10_13_indexes(db *sql.DB) {
	name := "migration_2023_10_13_indexes"

	if isMigrationExists(db, name) {
		return
	}

	_, err := db.Exec(`
		DROP PROCEDURE IF EXISTS drop_indexes;
		CREATE PROCEDURE drop_indexes()
		BEGIN
			IF EXISTS (SELECT * FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_NAME = 'session'
				AND INDEX_NAME = 'idx_session_completed_at_is_completed' AND INDEX_SCHEMA='stats') THEN
			ALTER TABLE session DROP INDEX idx_session_completed_at_is_completed;
			END IF;
		END;

		CALL drop_indexes();
		DROP PROCEDURE IF EXISTS drop_indexes;

		CREATE INDEX idx_session_completed_at_is_completed ON session ((date(started_at)), is_completed);
		`,
	)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
