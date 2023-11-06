package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
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

			is_completed BOOLEAN GENERATED ALWAYS AS (status IN (2,3,4)) STORED,

			FOREIGN KEY (server_id) REFERENCES server(id) ON UPDATE CASCADE ON DELETE CASCADE,
			FOREIGN KEY (map_id) REFERENCES maps(id) ON UPDATE CASCADE ON DELETE CASCADE,

			INDEX idx_session_is_completed_completed_at (is_completed, (date(completed_at)))
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
		CREATE TABLE IF NOT EXISTS session_game_data_cd (
			session_id INTEGER PRIMARY KEY NOT NULL,
			
			spawn_cycle TEXT NOT NULL,
			max_monsters INTEGER NOT NULL,
			wave_size_fakes INTEGER NOT NULL,
			zeds_type TEXT NOT NULL,

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
		CREATE TABLE IF NOT EXISTS wave_stats_cd (
			stats_id INTEGER PRIMARY KEY NOT NULL,

			spawn_cycle TEXT NOT NULL,
			max_monsters INTEGER NOT NULL,
			wave_size_fakes INTEGER NOT NULL,
			zeds_type TEXT NOT NULL,

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
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS wave_stats_player_injured_by (
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
			husk INTEGER NOT NULL,

			scrake INTEGER NOT NULL,
			fp INTEGER NOT NULL,
			qp INTEGER NOT NULL,
			boss INTEGER NOT NULL,

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
			
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
			FOREIGN KEY (current_session_id) REFERENCES session(id) ON UPDATE SET NULL ON DELETE SET NULL,
			FOREIGN KEY (last_session_id) REFERENCES session(id) ON UPDATE SET NULL ON DELETE SET NULL,

			INDEX idx_users_activity_curr (current_session_id)
		)`,
	)
	tx.Exec(`
		DROP TRIGGER IF EXISTS insert_aggregated_kills;
		CREATE TRIGGER insert_aggregated_kills
		AFTER INSERT ON wave_stats_player_kills
		FOR EACH ROW
		BEGIN
			INSERT INTO aggregated_kills (player_stats_id, trash, medium, large, total)
			VALUES (
				new.player_stats_id,
				new.cyst + new.alpha_clot + new.slasher + 
				new.stalker + new.crawler + new.gorefast + 
				new.rioter + new.elite_crawler + new.gorefiend,
				new.siren + new.bloat + new.edar + new.husk_n + new.husk_b, 
				new.scrake + new.fp + new.qp, 
				new.cyst + new.alpha_clot + new.slasher + 
				new.stalker + new.crawler + new.gorefast + 
				new.rioter + new.elite_crawler + new.gorefiend + 
				new.siren + new.bloat + new.edar + new.husk_n + new.husk_b + 
				new.scrake + new.fp + new.qp + new.boss + new.custom
			);
		END;
	`)
	tx.Exec(`
		DROP TRIGGER IF EXISTS update_user_activity_on_wave_end;
		CREATE TRIGGER update_user_activity_on_wave_end
		AFTER INSERT ON wave_stats_player
		FOR EACH ROW
		BEGIN
			UPDATE users_activity
			SET current_session_id = 
				(select min(ws.session_id) from wave_stats ws
				inner join wave_stats_player wsp on wsp.stats_id = ws.id
				where wsp.id = new.id),
				updated_at = CURRENT_TIMESTAMP
			WHERE user_id = new.player_id;
		END;
	`)
	tx.Exec(`
		DROP TRIGGER IF EXISTS update_user_activity_on_session_end;
		CREATE TRIGGER update_user_activity_on_session_end
		AFTER UPDATE ON session
		FOR EACH ROW
		BEGIN
			IF new.status <> old.status && new.status IN (2,3,4,-1) THEN
				UPDATE users_activity
				SET last_session_id = current_session_id, 
					current_session_id = NULL, 
					updated_at = CURRENT_TIMESTAMP
				WHERE current_session_id = new.id;
			END IF;
		END;
	`)

	return err
}

func NewDBInstance(user, pass, host, db string, port int) *sql.DB {
	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local&multiStatements=True",
		user, pass, host, port, db,
	)

	instance, err := sql.Open("mysql", connString)
	if err != nil {
		panic(err)
	}

	if err := initSchema(instance); err != nil {
		panic(err)
	}

	return instance
}
