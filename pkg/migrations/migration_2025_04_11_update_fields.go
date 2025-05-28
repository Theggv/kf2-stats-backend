package migrations

import "database/sql"

func migration_2025_04_11_update_fields(db *sql.DB) {
	name := "migration_2025_04_11_update_fields"

	if isMigrationExists(db, name) {
		return
	}

	_, err := db.Exec(`
 		DROP PROCEDURE IF EXISTS migration_2025_04_11_update_fields;
 		CREATE PROCEDURE migration_2025_04_11_update_fields()
 		BEGIN
 			IF EXISTS (
				SELECT * 
				FROM INFORMATION_SCHEMA.STATISTICS 
				WHERE TABLE_NAME = 'session_game_data_extra'
			) THEN
			 	ALTER TABLE session_game_data_extra MODIFY COLUMN spawn_cycle TEXT;
			 	ALTER TABLE session_game_data_extra MODIFY COLUMN max_monsters INTEGER;
			 	ALTER TABLE session_game_data_extra MODIFY COLUMN wave_size_fakes INTEGER;
			 	ALTER TABLE session_game_data_extra MODIFY COLUMN zeds_type TEXT;
			 	ALTER TABLE session_game_data_extra ADD percentage INTEGER AFTER zeds_type;
			 	ALTER TABLE session_game_data_extra ADD extra_percentage INTEGER AFTER percentage;
 			END IF;

			IF EXISTS (
				SELECT * 
				FROM INFORMATION_SCHEMA.STATISTICS 
				WHERE TABLE_NAME = 'wave_stats_extra'
			) THEN
			 	ALTER TABLE wave_stats_extra MODIFY COLUMN spawn_cycle TEXT;
			 	ALTER TABLE wave_stats_extra MODIFY COLUMN max_monsters INTEGER;
			 	ALTER TABLE wave_stats_extra MODIFY COLUMN wave_size_fakes INTEGER;
			 	ALTER TABLE wave_stats_extra MODIFY COLUMN zeds_type TEXT;
			 	ALTER TABLE wave_stats_extra ADD percentage INTEGER AFTER zeds_type;
			 	ALTER TABLE wave_stats_extra ADD extra_percentage INTEGER AFTER percentage;
 			END IF;
 		END;
 
 		CALL migration_2025_04_11_update_fields();
 		DROP PROCEDURE IF EXISTS migration_2025_04_11_update_fields;
 		`,
	)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
