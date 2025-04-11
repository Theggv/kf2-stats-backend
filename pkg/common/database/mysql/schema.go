package mysql

import (
	"context"
	"database/sql"
)

func initSchema(db *sql.DB) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	tx.Exec(`
		CREATE TABLE IF NOT EXISTS maps (
			id INTEGER PRIMARY KEY AUTO_INCREMENT, 
			name VARCHAR(64) NOT NULL, 
			preview TEXT,
			
			UNIQUE INDEX idx_uniq_maps_name (name) 
		)`,
	)
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS server (
			id INTEGER PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(128) NOT NULL, 
			address VARCHAR(64) NOT NULL,
		
			UNIQUE INDEX idx_uniq_server_address (address)
		)`,
	)
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTO_INCREMENT,
			auth_id VARCHAR(64) NOT NULL,
			auth_type INTEGER NOT NULL,
		
			name VARCHAR(64) NOT NULL,
			
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
		
			UNIQUE INDEX idx_uniq_users_auth (auth_id, auth_type)
		)`,
	)
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS session (
			id INTEGER PRIMARY KEY AUTO_INCREMENT,
			server_id INTEGER NOT NULL,
			map_id INTEGER NOT NULL,

			mode INTEGER NOT NULL,
			length INTEGER NOT NULL,
			diff INTEGER NOT NULL,

			status INTEGER NOT NULL DEFAULT 0,

			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,

			is_completed BOOLEAN GENERATED ALWAYS AS (status IN (-1,2,3,4)) STORED,

			FOREIGN KEY (server_id) REFERENCES server(id) ON UPDATE CASCADE ON DELETE CASCADE,
			FOREIGN KEY (map_id) REFERENCES maps(id) ON UPDATE CASCADE ON DELETE CASCADE,

			INDEX idx_session_completed_at_is_completed ((date(started_at)), is_completed)
		)
	`)
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS session_demo (
			session_id INTEGER PRIMARY KEY NOT NULL,

			data LONGBLOB NOT NULL,
			processed BOOLEAN NOT NULL DEFAULT 0,

			FOREIGN KEY (session_id) REFERENCES session(id) ON UPDATE CASCADE ON DELETE CASCADE
		)
	`)
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS session_game_data (
			session_id INTEGER PRIMARY KEY NOT NULL,

			max_players INTEGER NOT NULL DEFAULT 6,
			players_online INTEGER NOT NULL DEFAULT 0,
			players_alive INTEGER NOT NULL DEFAULT 0,
			
			wave INTEGER NOT NULL DEFAULT 0,
			is_trader_time BOOLEAN NOT NULL DEFAULT 0,
			zeds_left INTEGER NOT NULL DEFAULT 0,

			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			FOREIGN KEY (session_id) REFERENCES session(id) ON UPDATE CASCADE ON DELETE CASCADE
		)
	`)
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS session_game_data_extra (
			session_id INTEGER PRIMARY KEY NOT NULL,
			
			spawn_cycle TEXT,
			max_monsters INTEGER,
			wave_size_fakes INTEGER,
			zeds_type TEXT,

			percentage INTEGER,
			extra_percentage INTEGER,

			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			FOREIGN KEY (session_id) REFERENCES session(id) ON UPDATE CASCADE ON DELETE CASCADE
		)
	`)
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS wave_stats (
			id INTEGER PRIMARY KEY AUTO_INCREMENT,
			session_id INTEGER NOT NULL,
			wave INTEGER NOT NULL,
			attempt INTEGER NOT NULL,

			started_at TIMESTAMP NOT NULL,
			completed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			FOREIGN KEY (session_id) REFERENCES session(id) ON UPDATE CASCADE ON DELETE CASCADE,

			UNIQUE INDEX idx_uniq_wave_stats (session_id, wave, attempt)
		)
	`)
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS wave_stats_extra (
			stats_id INTEGER PRIMARY KEY NOT NULL,

			spawn_cycle TEXT,
			max_monsters INTEGER,
			wave_size_fakes INTEGER,
			zeds_type TEXT,

			percentage INTEGER,
			extra_percentage INTEGER,

			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			FOREIGN KEY (stats_id) REFERENCES wave_stats(id) ON UPDATE CASCADE ON DELETE CASCADE
		)
	`)
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS wave_stats_player (
			id INTEGER PRIMARY KEY AUTO_INCREMENT,
			stats_id INTEGER NOT NULL,
			player_id INTEGER NOT NULL,

			perk INTEGER NOT NULL,
			level INTEGER NOT NULL,
			prestige INTEGER NOT NULL,

			is_dead BOOLEAN NOT NULL,

			shots_fired INTEGER NOT NULL,
			shots_hit INTEGER NOT NULL,
			shots_hs INTEGER NOT NULL,

			dosh_earned INTEGER NOT NULL,

			heals_given INTEGER NOT NULL,
			heals_recv INTEGER NOT NULL,

			damage_dealt INTEGER NOT NULL,
			damage_taken INTEGER NOT NULL,

			zedtime_count INTEGER NOT NULL,
			zedtime_length REAL NOT NULL,

			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			FOREIGN KEY (stats_id) REFERENCES wave_stats(id) ON UPDATE CASCADE ON DELETE CASCADE,
			FOREIGN KEY (player_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,

			UNIQUE INDEX idx_uniq_wave_stats_player (stats_id, player_id),
			INDEX idx_wave_stats_player_player_id (player_id)
		)
	`)
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS wave_stats_player_kills (
			player_stats_id INTEGER PRIMARY KEY NOT NULL,

			cyst INTEGER NOT NULL,
			alpha_clot INTEGER NOT NULL,
			slasher INTEGER NOT NULL,
			stalker INTEGER NOT NULL,
			crawler INTEGER NOT NULL,
			gorefast INTEGER NOT NULL,
			rioter INTEGER NOT NULL,
			elite_crawler INTEGER NOT NULL,
			gorefiend INTEGER NOT NULL,

			siren INTEGER NOT NULL,
			bloat INTEGER NOT NULL,
			edar INTEGER NOT NULL,
			husk_n INTEGER NOT NULL,
			husk_b INTEGER NOT NULL,
			husk_r INTEGER NOT NULL,

			scrake INTEGER NOT NULL,
			fp INTEGER NOT NULL,
			qp INTEGER NOT NULL,
			boss INTEGER NOT NULL,
			custom INTEGER NOT NULL,

			FOREIGN KEY (player_stats_id) REFERENCES wave_stats_player(id) ON UPDATE CASCADE ON DELETE CASCADE
		)
	`)
	tx.Exec(
		`CREATE TABLE IF NOT EXISTS wave_stats_player_comms (
			player_stats_id INTEGER PRIMARY KEY NOT NULL,

			request_healing INTEGER NOT NULL,
			request_dosh INTEGER NOT NULL,
			request_help INTEGER NOT NULL,
			taunt_zeds INTEGER NOT NULL,
			follow_me INTEGER NOT NULL,
			get_to_the_trader INTEGER NOT NULL,
			affirmative INTEGER NOT NULL,
			negative INTEGER NOT NULL,
			thank_you INTEGER NOT NULL,

			FOREIGN KEY (player_stats_id) REFERENCES wave_stats_player(id) ON UPDATE CASCADE ON DELETE CASCADE
		)
	`)
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS aggregated_kills (
			player_stats_id INTEGER PRIMARY KEY NOT NULL,

			trash INTEGER NOT NULL,
			medium INTEGER NOT NULL,
			large INTEGER NOT NULL,
			total INTEGER NOT NULL,

			FOREIGN KEY (player_stats_id) REFERENCES wave_stats_player(id) ON UPDATE CASCADE ON DELETE CASCADE
		)
	`)
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS users_activity (
			user_id INTEGER PRIMARY KEY NOT NULL,
			current_session_id INTEGER,
			last_session_id INTEGER,

			perk INTEGER NOT NULL DEFAULT 0,
			level INTEGER NOT NULL DEFAULT 0,
			prestige INTEGER NOT NULL DEFAULT 0,

			health INTEGER NOT NULL DEFAULT 0,
			armor INTEGER NOT NULL DEFAULT 0,
			is_spectator BOOLEAN NOT NULL DEFAULT FALSE,
			
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
			FOREIGN KEY (current_session_id) REFERENCES session(id) ON UPDATE SET NULL ON DELETE SET NULL,
			FOREIGN KEY (last_session_id) REFERENCES session(id) ON UPDATE SET NULL ON DELETE SET NULL,

			INDEX idx_users_activity_curr (current_session_id)
		)`,
	)
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS session_aggregated (
			id INTEGER PRIMARY KEY AUTO_INCREMENT,
			session_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			perk INTEGER NOT NULL,

			playtime_seconds INTEGER NOT NULL,
			waves_played INTEGER NOT NULL, 
			deaths INTEGER NOT NULL,

			shots_fired INTEGER NOT NULL,
			shots_hit INTEGER NOT NULL,
			shots_hs INTEGER NOT NULL,

			dosh_earned INTEGER NOT NULL,

			heals_given INTEGER NOT NULL,
			heals_recv INTEGER NOT NULL,

			damage_dealt INTEGER NOT NULL,
			damage_taken INTEGER NOT NULL,

			zedtime_count INTEGER NOT NULL,
			zedtime_length REAL NOT NULL,

			FOREIGN KEY (session_id) REFERENCES session(id) ON UPDATE CASCADE ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,

			UNIQUE INDEX idx_uniq_perk_user_id_session_id (perk, user_id, session_id)
		)
	`)
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS session_aggregated_kills (
			id INTEGER PRIMARY KEY NOT NULL,

			trash INTEGER NOT NULL,
			medium INTEGER NOT NULL,
			large INTEGER NOT NULL,
			total INTEGER NOT NULL,

			FOREIGN KEY (id) REFERENCES session_aggregated(id) ON UPDATE CASCADE ON DELETE CASCADE
		)
	`)
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS user_weekly_stats_common (
			period INTEGER NOT NULL,
			user_id INTEGER NOT NULL,

			total_games INTEGER NOT NULL,
			total_waves INTEGER NOT NULL, 
			playtime_seconds INTEGER NOT NULL,
			deaths INTEGER NOT NULL,

			shots_fired INTEGER NOT NULL,
			shots_hit INTEGER NOT NULL,
			shots_hs INTEGER NOT NULL,

			dosh_earned INTEGER NOT NULL,

			heals_given INTEGER NOT NULL,
			heals_recv INTEGER NOT NULL,

			damage_dealt INTEGER NOT NULL,
			damage_taken INTEGER NOT NULL,

			max_damage_session_id INTEGER NOT NULL,
			max_damage INTEGER NOT NULL,

			PRIMARY KEY (period, user_id),

			FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE
		)
	`)

	tx.Exec(`
		CREATE TABLE IF NOT EXISTS user_weekly_stats_perk (
			period INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			perk INTEGER NOT NULL,

			total_games INTEGER NOT NULL,
			total_waves INTEGER NOT NULL,
			playtime_seconds INTEGER NOT NULL,
			deaths INTEGER NOT NULL,

			shots_fired INTEGER NOT NULL,
			shots_hit INTEGER NOT NULL,
			shots_hs INTEGER NOT NULL,

			dosh_earned INTEGER NOT NULL,

			heals_given INTEGER NOT NULL,
			heals_recv INTEGER NOT NULL,

			damage_dealt INTEGER NOT NULL,
			damage_taken INTEGER NOT NULL,

			zedtime_count INTEGER NOT NULL,
			zedtime_length REAL NOT NULL,
			
			buffs_active_length REAL NOT NULL,
			buffs_total_length REAL NOT NULL,

			max_damage_session_id INTEGER NOT NULL,
			max_damage INTEGER NOT NULL,

			PRIMARY KEY (period, user_id, perk),

			FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE
		)
	`)

	return err
}
