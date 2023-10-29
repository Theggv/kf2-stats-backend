package migrations

import "database/sql"

func migration_27_10_2023_fix_game_data_cd(db *sql.DB) {
	name := "migration_27_10_2023_fix_game_data_cd"

	if isMigrationExists(db, name) {
		return
	}

	_, err := db.Exec(`
		UPDATE session_game_data_cd
		SET 
			spawn_cycle = new.spawn_cycle, 
			max_monsters = new.max_monsters, 
			wave_size_fakes = new.wave_size_fakes, 
			zeds_type = new.zeds_type
		FROM (SELECT 
			session.id as session_id,
			max(stats_id),
			cd.spawn_cycle as spawn_cycle,
			cd.max_monsters as max_monsters,
			cd.wave_size_fakes as wave_size_fakes,
			cd.zeds_type as zeds_type
		FROM session
		INNER JOIN wave_stats ws ON ws.session_id = session.id
		INNER JOIN wave_stats_cd cd ON cd.stats_id = ws.id
		WHERE ws.wave <= session.length
		GROUP BY session.id) as new
		WHERE session_game_data_cd.session_id = new.session_id`,
	)

	if err != nil {
		panic(err)
	}

	writeMigration(db, name)
}
