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
				SET buffs_active_length = new.buffs_active_length, 
					buffs_total_length = new.buffs_total_length
				WHERE session.id = old.session_id;
			END IF;
		END;
	`)
	tx.Exec(`
		DROP TRIGGER IF EXISTS update_user_weekly_stats_total_max_damage;
		CREATE TRIGGER update_user_weekly_stats_total_max_damage
		BEFORE UPDATE ON user_weekly_stats_total
		FOR EACH ROW
		BEGIN
			IF new.total_games > 0 THEN
				SET new.total_games = old.total_games + new.total_games;
			END IF;
			IF new.total_waves > 0 THEN
				SET new.total_waves = old.total_waves + new.total_waves; 
			END IF;
			IF new.playtime_seconds > 0 THEN
				SET new.playtime_seconds = old.playtime_seconds + new.playtime_seconds; 
			END IF;
			IF new.deaths > 0 THEN
				SET new.deaths = old.deaths + new.deaths; 
			END IF;

			IF new.shots_fired > 0 THEN
				SET new.shots_fired = old.shots_fired + new.shots_fired; 
			END IF;
			IF new.shots_hit > 0 THEN
				SET new.shots_hit = old.shots_hit + new.shots_hit; 
			END IF;
			IF new.shots_hs > 0 THEN
				SET new.shots_hs = old.shots_hs + new.shots_hs; 
			END IF;

			IF new.dosh_earned > 0 THEN
				SET new.dosh_earned = old.dosh_earned + new.dosh_earned; 
			END IF;	
			IF new.heals_given > 0 THEN
				SET new.heals_given = old.heals_given + new.heals_given; 
			END IF;	
			IF new.heals_recv > 0 THEN
				SET new.heals_recv = old.heals_recv + new.heals_recv; 
			END IF;	

			IF new.damage_dealt > 0 THEN
				SET new.damage_dealt = old.damage_dealt + new.damage_dealt; 
			END IF;	
			IF new.damage_taken > 0 THEN
				SET new.damage_taken = old.damage_taken + new.damage_taken;
			END IF;	

			IF new.large_kills > 0 THEN
				SET new.large_kills = old.large_kills + new.large_kills; 
			END IF;	
			IF new.total_kills > 0 THEN
				SET new.total_kills = old.total_kills + new.total_kills;
			END IF;	

			IF new.max_damage < old.max_damage THEN
				SET new.max_damage = old.max_damage;
				SET new.max_damage_session_id = old.max_damage_session_id;
			END IF;
		END;
	`)
	tx.Exec(`
		DROP TRIGGER IF EXISTS update_user_weekly_stats_perk_max_damage;
		CREATE TRIGGER update_user_weekly_stats_perk_max_damage
		BEFORE UPDATE ON user_weekly_stats_perk
		FOR EACH ROW
		BEGIN
			IF new.total_games > 0 THEN
				SET new.total_games = old.total_games + new.total_games;
			END IF;
			IF new.total_waves > 0 THEN
				SET new.total_waves = old.total_waves + new.total_waves; 
			END IF;
			IF new.playtime_seconds > 0 THEN
				SET new.playtime_seconds = old.playtime_seconds + new.playtime_seconds; 
			END IF;
			IF new.deaths > 0 THEN
				SET new.deaths = old.deaths + new.deaths; 
			END IF;

			IF new.shots_fired > 0 THEN
				SET new.shots_fired = old.shots_fired + new.shots_fired; 
			END IF;
			IF new.shots_hit > 0 THEN
				SET new.shots_hit = old.shots_hit + new.shots_hit; 
			END IF;
			IF new.shots_hs > 0 THEN
				SET new.shots_hs = old.shots_hs + new.shots_hs; 
			END IF;

			IF new.dosh_earned > 0 THEN
				SET new.dosh_earned = old.dosh_earned + new.dosh_earned; 
			END IF;	
			IF new.heals_given > 0 THEN
				SET new.heals_given = old.heals_given + new.heals_given; 
			END IF;	
			IF new.heals_recv > 0 THEN
				SET new.heals_recv = old.heals_recv + new.heals_recv; 
			END IF;	

			IF new.damage_dealt > 0 THEN
				SET new.damage_dealt = old.damage_dealt + new.damage_dealt; 
			END IF;	
			IF new.damage_taken > 0 THEN
				SET new.damage_taken = old.damage_taken + new.damage_taken;
			END IF;	

			IF new.large_kills > 0 THEN
				SET new.large_kills = old.large_kills + new.large_kills; 
			END IF;	
			IF new.total_kills > 0 THEN
				SET new.total_kills = old.total_kills + new.total_kills;
			END IF;

			IF new.zedtime_count > 0 THEN
				SET new.zedtime_count = old.zedtime_count + new.zedtime_count; 
			END IF;	
			IF new.zedtime_length > 0 THEN
				SET new.zedtime_length = old.zedtime_length + new.zedtime_length;
			END IF;	

			IF new.buffs_active_length > 0 THEN
				SET new.buffs_active_length = old.buffs_active_length + new.buffs_active_length; 
			END IF;	
			IF new.buffs_total_length > 0 THEN
				SET new.buffs_total_length = old.buffs_total_length + new.buffs_total_length;
			END IF;	

			IF new.max_damage < old.max_damage THEN
				SET new.max_damage = old.max_damage;
				SET new.max_damage_session_id = old.max_damage_session_id;
			END IF;
		END;
	`)

	return err
}
