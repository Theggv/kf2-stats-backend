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
				SET weekly.buffs_active_length = weekly.buffs_active_length + new.buffs_active_length, 
					weekly.buffs_total_length = weekly.buffs_total_length + new.buffs_total_length
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
			IF new.max_damage < old.max_damage THEN
				SET new.max_damage = old.max_damage;
				SET new.max_damage_session_id = old.max_damage_session_id;
			END IF;
		END;
	`)

	return err
}
