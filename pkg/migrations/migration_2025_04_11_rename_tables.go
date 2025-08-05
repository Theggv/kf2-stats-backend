package migrations

import "database/sql"

func migration_2025_04_11_rename_tables(db *sql.DB) {
	name := "migration_2025_04_11_rename_tables"

	if isMigrationExists(db, name) {
		return
	}

	_, err := db.Exec(`
 		DROP PROCEDURE IF EXISTS migration_2025_04_11_rename_tables;
 		CREATE PROCEDURE migration_2025_04_11_rename_tables()
 		BEGIN
 			IF EXISTS (
				SELECT * 
				FROM INFORMATION_SCHEMA.STATISTICS 
				WHERE TABLE_NAME = 'session_game_data_cd'
			) AND EXISTS (
				SELECT * 
				FROM INFORMATION_SCHEMA.STATISTICS 
				WHERE TABLE_NAME = 'session_game_data_extra'
			) THEN
				DROP TABLE session_game_data_extra;
 				RENAME TABLE session_game_data_cd TO session_game_data_extra;
 			END IF;

			IF EXISTS (
				SELECT * 
				FROM INFORMATION_SCHEMA.STATISTICS 
				WHERE TABLE_NAME = 'wave_stats_cd'
			) AND EXISTS (
				SELECT * 
				FROM INFORMATION_SCHEMA.STATISTICS 
				WHERE TABLE_NAME = 'wave_stats_extra'
			) THEN
				DROP TABLE wave_stats_extra;
 				RENAME TABLE wave_stats_cd TO wave_stats_extra;
 			END IF;
 		END;
 
 		CALL migration_2025_04_11_rename_tables();
 		DROP PROCEDURE IF EXISTS migration_2025_04_11_rename_tables;
 		`,
	)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
