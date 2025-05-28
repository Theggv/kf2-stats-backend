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

			DROP TRIGGER IF EXISTS update_session_aggregated_post;
			CREATE TRIGGER update_session_aggregated_post
			AFTER UPDATE ON session_aggregated
			FOR EACH ROW
			BEGIN
				IF new.buffs_active_length <> old.buffs_active_length && new.buffs_active_length > 0 THEN
					UPDATE user_weekly_stats_perk weekly
					INNER JOIN session ON 
						weekly.period = yearweek(session.started_at) AND
						weekly.server_id = session.server_id AND
						weekly.perk = old.perk AND
						weekly.user_id = old.user_id
					SET weekly.buffs_active_length = weekly.buffs_active_length + new.buffs_active_length, 
						weekly.buffs_total_length = weekly.buffs_total_length + new.buffs_total_length
					WHERE session.id = old.session_id;
				END IF;
			END;
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
