package mysql

import (
	"context"
	"database/sql"
)

func initTriggers(db *sql.DB) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

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

				CALL insert_session_aggregated(new.id);
			END IF;
		END;
	`)
	tx.Exec(`
		DROP TRIGGER IF EXISTS update_user_weekly_stats_perk_max_damage;
		CREATE TRIGGER update_user_weekly_stats_perk_max_damage
		BEFORE UPDATE ON user_weekly_stats_perk
		FOR EACH ROW
		BEGIN
			SET new.total_games = old.total_games + new.total_games; 
			SET new.total_waves = old.total_waves + new.total_waves; 
			SET new.playtime_seconds = old.playtime_seconds + new.playtime_seconds; 
			SET new.deaths = old.deaths + new.deaths; 

			SET new.shots_fired = old.shots_fired + new.shots_fired; 
			SET new.shots_hit = old.shots_hit + new.shots_hit; 
			SET new.shots_hs = old.shots_hs + new.shots_hs; 

			SET new.dosh_earned = old.dosh_earned + new.dosh_earned; 
			SET new.heals_given = old.heals_given + new.heals_given; 
			SET new.heals_recv = old.heals_recv + new.heals_recv; 

			SET new.damage_dealt = old.damage_dealt + new.damage_dealt; 
			SET new.damage_taken = old.damage_taken + new.damage_taken; 

			SET new.zedtime_count = old.zedtime_count + new.zedtime_count; 
			SET new.zedtime_length = old.zedtime_length + new.zedtime_length;

			SET new.buffs_active_length = old.buffs_active_length + new.buffs_active_length; 
			SET new.buffs_total_length = old.buffs_total_length + new.buffs_total_length;
			
			IF new.max_damage < old.max_damage THEN
				SET new.max_damage = old.max_damage;
				SET new.max_damage_session_id = old.max_damage_session_id;
			END IF;
		END;
	`)
	tx.Exec(`
		DROP TRIGGER IF EXISTS update_user_weekly_stats_common_max_damage;
		CREATE TRIGGER update_user_weekly_stats_common_max_damage
		BEFORE UPDATE ON user_weekly_stats_common
		FOR EACH ROW
		BEGIN
			SET new.total_games = old.total_games + new.total_games; 
			SET new.total_waves = old.total_waves + new.total_waves; 
			SET new.playtime_seconds = old.playtime_seconds + new.playtime_seconds; 
			SET new.deaths = old.deaths + new.deaths; 

			SET new.shots_fired = old.shots_fired + new.shots_fired; 
			SET new.shots_hit = old.shots_hit + new.shots_hit; 
			SET new.shots_hs = old.shots_hs + new.shots_hs; 

			SET new.dosh_earned = old.dosh_earned + new.dosh_earned; 
			SET new.heals_given = old.heals_given + new.heals_given; 
			SET new.heals_recv = old.heals_recv + new.heals_recv; 

			SET new.damage_dealt = old.damage_dealt + new.damage_dealt; 
			SET new.damage_taken = old.damage_taken + new.damage_taken; 

			IF new.max_damage < old.max_damage THEN
				SET new.max_damage = old.max_damage;
				SET new.max_damage_session_id = old.max_damage_session_id;
			END IF;
		END;
	`)

	return err
}
