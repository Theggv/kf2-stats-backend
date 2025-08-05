package mysql

import (
	"context"
	"database/sql"
)

func initStored(db *sql.DB) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	tx.Exec(`
		DROP PROCEDURE IF EXISTS fix_dropped_sessions;
		CREATE PROCEDURE fix_dropped_sessions()
		BEGIN
			UPDATE session
			INNER JOIN (
				SELECT server_id, max(id) as max_id FROM session
				GROUP BY server_id
			) as tbl on session.server_id = tbl.server_id
			SET status = -1
			WHERE 
				session.id NOT IN (tbl.max_id) AND 
				status IN (0, 1);
		END;
	`)
	tx.Exec(`
		DROP PROCEDURE IF EXISTS abort_old_sessions;
		CREATE PROCEDURE abort_old_sessions(IN minutes int)
		BEGIN
			UPDATE session
			INNER JOIN server ON server.id = session.server_id
			INNER JOIN session_game_data gd ON gd.session_id = session.id
			SET session.status = -1
			WHERE 
				session.status IN (0, 1) AND 
				session.started_at >= CURRENT_TIMESTAMP - INTERVAL 30 DAY AND
				timestampdiff(MINUTE, gd.updated_at, CURRENT_TIMESTAMP) > minutes;
		END;
	`)
	tx.Exec(`
		DROP PROCEDURE IF EXISTS delete_empty_sessions;
		CREATE PROCEDURE delete_empty_sessions()
		BEGIN
			DELETE FROM session 
			WHERE
				session.status IN (-1, 2, 3) AND
				session.started_at >= CURRENT_TIMESTAMP - INTERVAL 30 DAY AND
				NOT EXISTS (
					SELECT 1
					FROM wave_stats ws
					INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
					WHERE ws.session_id = session.id
				);
		END;
	`)
	tx.Exec(`
		DROP FUNCTION IF EXISTS get_user_games_by_perk;
		CREATE FUNCTION get_user_games_by_perk(user_id INT, perk INT, date_from DATE, date_to DATE)
		RETURNS INTEGER READS SQL DATA
		BEGIN
			DECLARE value INTEGER;

			SELECT count(session.id) INTO value
			FROM session
			INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
			WHERE 
				aggr.user_id = user_id AND 
				aggr.perk = perk AND 
				session.started_at BETWEEN date_from AND date_to;

			RETURN value;
		END;
	`)
	tx.Exec(`
		DROP FUNCTION IF EXISTS get_avg_zt;
		-- Get user's average zed time length for certain period by using trimmed mean.
		CREATE FUNCTION get_avg_zt(user_id INT, date_from DATE, date_to DATE)
		RETURNS REAL READS SQL DATA
		BEGIN
			DECLARE total_games INTEGER;
			DECLARE percent REAL;
			DECLARE value REAL;
			
			SELECT get_user_games_by_perk(user_id, 2, date_from, date_to) INTO total_games;
		
			CASE
				WHEN total_games < 10 THEN SET percent = 0;
				ELSE SET percent = 0.1;
			END CASE;
			
			SELECT coalesce(avg(avg_zedtime), 0) INTO value
			FROM (
				SELECT 
					avg_zedtime,
					cume_dist() OVER (ORDER BY avg_zedtime) as dist 
				FROM (
					SELECT round(zedtime_length / greatest(zedtime_count, 1), 2) as avg_zedtime
					FROM session
					INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
					WHERE 
						aggr.user_id = user_id AND 
						aggr.perk = 2 AND
						aggr.playtime_seconds >= 30 AND
						session.started_at BETWEEN date_from AND date_to
					GROUP BY aggr.id
				) t
			) t
			WHERE dist >= percent AND dist <= (1 - percent);

			RETURN round(value, 2);
		END;
	`)
	tx.Exec(`
		DROP FUNCTION IF EXISTS get_avg_acc;
		CREATE FUNCTION get_avg_acc(user_id INT, perk INT, date_from DATE, date_to DATE)
		RETURNS REAL READS SQL DATA
		BEGIN
			DECLARE total_games INTEGER;
			DECLARE percent REAL;
			DECLARE value REAL;
			
			SELECT get_user_games_by_perk(user_id, perk, date_from, date_to) INTO total_games;
		
			CASE
				WHEN total_games < 10 THEN SET percent = 0;
				ELSE SET percent = 0.1;
			END CASE;
			
			SELECT coalesce(sum(shots_hit) / greatest(sum(shots_fired), 1), 0) INTO value
			FROM (
				SELECT 
					shots_hit,
					shots_fired,
					cume_dist() OVER (ORDER BY accuracy) as dist
				FROM (
					SELECT
						min(shots_hit) as shots_hit,
						min(shots_fired) as shots_fired,
						min(shots_hit) / greatest(min(shots_fired), 1) as accuracy
					FROM session
					INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
					WHERE 
						aggr.user_id = user_id AND 
						aggr.perk = perk AND 
						aggr.playtime_seconds >= 30 AND
						session.started_at BETWEEN date_from AND date_to
					GROUP BY session.id
				) t
			) t
			WHERE dist >= percent AND dist <= (1 - percent);

			RETURN round(value, 2);
		END;
	`)
	tx.Exec(`
		DROP FUNCTION IF EXISTS get_avg_hs_acc;
		CREATE FUNCTION get_avg_hs_acc(user_id INT, perk INT, date_from DATE, date_to DATE)
		RETURNS REAL READS SQL DATA
		BEGIN
			DECLARE total_games INTEGER;
			DECLARE percent REAL;
			DECLARE value REAL;
			
			SELECT get_user_games_by_perk(user_id, perk, date_from, date_to) INTO total_games;
		
			CASE
				WHEN total_games < 10 THEN SET percent = 0;
				ELSE SET percent = 0.1;
			END CASE;
			
			SELECT coalesce(sum(shots_hs) / greatest(sum(shots_hit), 1), 0) INTO value
			FROM (
				SELECT 
					shots_hit,
					shots_hs,
					cume_dist() OVER (ORDER BY hs_acc) as dist
				FROM (
					SELECT
						min(shots_hit) as shots_hit,
						min(shots_hs) as shots_hs,
						min(shots_hs) / greatest(min(shots_hit), 1) as hs_acc
					FROM session
					INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
					WHERE 
						aggr.user_id = user_id AND 
						aggr.perk = perk AND 
						aggr.playtime_seconds >= 30 AND
						session.started_at BETWEEN date_from AND date_to
					GROUP BY aggr.id
				) t
			) t
			WHERE dist >= percent AND dist <= (1 - percent);

			RETURN round(value, 2);
		END;
	`)
	tx.Exec(`
		DROP FUNCTION IF EXISTS get_user_games_count_by_perk;
		CREATE FUNCTION get_user_games_count_by_perk(user_id INT, perk INT, status_id INT, date_from DATE, date_to DATE)
		RETURNS INTEGER READS SQL DATA
		BEGIN
			DECLARE value INTEGER;

			SELECT count(t.status) INTO value
			FROM (
				SELECT status
				FROM session
				INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
				WHERE 
					aggr.user_id = user_id AND 
					aggr.perk = perk AND 
					session.started_at BETWEEN date_from AND date_to
				GROUP BY session.id
			) t
			WHERE status = status_id;

			RETURN value;
		END;
	`)
	tx.Exec(`
		DROP PROCEDURE IF EXISTS insert_session_aggregated;
		CREATE PROCEDURE insert_session_aggregated(session_id INT)
		BEGIN
			INSERT IGNORE INTO session_aggregated (
				session_id, user_id, perk, 
				playtime_seconds, waves_played, deaths, 
				shots_fired, shots_hit, shots_hs, 
				dosh_earned, heals_given, heals_recv, 
				damage_dealt, damage_taken, 
				zedtime_count, zedtime_length)
			(
				SELECT 
					session.id as session_id, 
					wsp.player_id as user_id, 
					wsp.perk as perk, 
					sum(timestampdiff(SECOND, ws.started_at, ws.completed_at)) as playtime_seconds,
					count(*) as waves_played,
					sum(is_dead = 1) as deaths, 
					sum(shots_fired) as shots_fired, 
					sum(shots_hit) as shots_hit, 
					sum(shots_hs) as shots_hs, 
					sum(dosh_earned) as dosh_earned, 
					sum(heals_given) as heals_given, 
					sum(heals_recv) as heals_recv, 
					sum(damage_dealt) as damage_dealt, 
					sum(damage_taken) as damage_taken, 
					sum(zedtime_count) as zedtime_count, 
					sum(zedtime_length) as zedtime_length
				FROM session
				INNER JOIN wave_stats ws ON ws.session_id = session.id
				INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
				WHERE session.id = session_id
				GROUP BY session.id, wsp.player_id, wsp.perk
			);

			INSERT IGNORE INTO session_aggregated_kills (id, trash, medium, large, total)
			(
				SELECT
					aggr.id,
					trash,
					medium,
					large,
					total
				FROM session_aggregated aggr
				INNER JOIN (
					SELECT 
						session.id as session_id, 
						wsp.player_id as user_id, 
						wsp.perk as perk, 
						sum(kills.trash) as trash, 
						sum(kills.medium) as medium, 
						sum(kills.large) as large, 
						sum(kills.total) as total
					FROM session
					INNER JOIN wave_stats ws ON ws.session_id = session.id
					INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
					INNER JOIN aggregated_kills kills ON kills.player_stats_id = wsp.id
					WHERE session.id = session_id
					GROUP BY session.id, wsp.player_id, wsp.perk
				) t ON aggr.session_id = t.session_id AND aggr.user_id = t.user_id AND aggr.perk = t.perk
			);

			CALL update_user_stats_weekly(session_id);
		END;
	`)
	tx.Exec(`
		DROP PROCEDURE IF EXISTS update_user_stats_weekly;
		CREATE PROCEDURE update_user_stats_weekly(session_id INT)
		BEGIN
			INSERT IGNORE INTO user_weekly_stats_total SELECT * FROM (
				SELECT
					YEARWEEK(session.started_at) as period, 

					session.server_id as server_id,
					wsp.player_id as user_id, 

					1 as total_games,
					count(*) as total_waves,
					sum(timestampdiff(SECOND, ws.started_at, ws.completed_at)) as playtime_seconds,
					sum(is_dead = 1) as deaths, 

					sum(shots_fired) as shots_fired, 
					sum(shots_hit) as shots_hit, 
					sum(shots_hs) as shots_hs, 

					sum(dosh_earned) as dosh_earned, 
					sum(heals_given) as heals_given, 
					sum(heals_recv) as heals_recv, 
					sum(damage_dealt) as damage_dealt, 
					sum(damage_taken) as damage_taken, 

					sum(k.scrake) + sum(k.fp) + sum(k.qp) as large_kills, 
					sum(k.cyst) + sum(k.alpha_clot) + sum(k.slasher) + 
					sum(k.stalker) + sum(k.crawler) + sum(k.gorefast) + 
					sum(k.rioter) + sum(k.elite_crawler) + sum(k.gorefiend) + 
					sum(k.siren) + sum(k.bloat) + sum(k.edar) + sum(k.husk_n) + sum(k.husk_b) + 
					sum(k.scrake) + sum(k.fp) + sum(k.qp) + sum(k.boss) + sum(k.custom) as total_kills,

					session.id as max_damage_session_id,
					sum(damage_dealt) as max_damage
				FROM session
				INNER JOIN wave_stats ws ON ws.session_id = session.id
				INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
				INNER JOIN wave_stats_player_kills k ON k.player_stats_id = wsp.id
				WHERE session.id = session_id
				GROUP BY session.id, wsp.player_id
			) as new
			ON DUPLICATE KEY UPDATE 
				user_weekly_stats_total.total_games = user_weekly_stats_total.total_games + new.total_games,
				user_weekly_stats_total.total_waves = user_weekly_stats_total.total_waves + new.total_waves,
				user_weekly_stats_total.playtime_seconds = user_weekly_stats_total.playtime_seconds + new.playtime_seconds,
				user_weekly_stats_total.deaths = user_weekly_stats_total.deaths + new.deaths,

				user_weekly_stats_total.shots_fired = user_weekly_stats_total.shots_fired + new.shots_fired,
				user_weekly_stats_total.shots_hit = user_weekly_stats_total.shots_hit + new.shots_hit,
				user_weekly_stats_total.shots_hs = user_weekly_stats_total.shots_hs + new.shots_hs,

				user_weekly_stats_total.dosh_earned = user_weekly_stats_total.dosh_earned + new.dosh_earned,
				user_weekly_stats_total.heals_given = user_weekly_stats_total.heals_given + new.heals_given,
				user_weekly_stats_total.heals_recv = user_weekly_stats_total.heals_recv + new.heals_recv,

				user_weekly_stats_total.damage_dealt = user_weekly_stats_total.damage_dealt + new.damage_dealt,
				user_weekly_stats_total.damage_taken = user_weekly_stats_total.damage_taken + new.damage_taken,
				
				user_weekly_stats_total.large_kills = user_weekly_stats_total.large_kills + new.large_kills,
				user_weekly_stats_total.total_kills = user_weekly_stats_total.total_kills + new.total_kills,

				max_damage_session_id = new.max_damage_session_id,
				max_damage = new.max_damage;

			INSERT IGNORE INTO user_weekly_stats_perk SELECT * FROM (
				SELECT
					YEARWEEK(session.started_at) as period, 

					session.server_id as server_id,
					wsp.player_id as user_id, 
					wsp.perk as perk,

					1 as total_games,
					count(*) as total_waves,
					sum(timestampdiff(SECOND, ws.started_at, ws.completed_at)) as playtime_seconds,
					sum(is_dead = 1) as deaths, 

					sum(shots_fired) as shots_fired, 
					sum(shots_hit) as shots_hit, 
					sum(shots_hs) as shots_hs, 

					sum(dosh_earned) as dosh_earned, 
					sum(heals_given) as heals_given, 
					sum(heals_recv) as heals_recv, 
					sum(damage_dealt) as damage_dealt, 
					sum(damage_taken) as damage_taken, 

					sum(zedtime_count) as zedtime_count, 
					sum(zedtime_length) as zedtime_length,

					0 as buffs_active_length, 
					0 as buffs_total_length,

					sum(k.scrake) + sum(k.fp) + sum(k.qp) as large_kills, 
					sum(k.cyst) + sum(k.alpha_clot) + sum(k.slasher) + 
					sum(k.stalker) + sum(k.crawler) + sum(k.gorefast) + 
					sum(k.rioter) + sum(k.elite_crawler) + sum(k.gorefiend) + 
					sum(k.siren) + sum(k.bloat) + sum(k.edar) + sum(k.husk_n) + sum(k.husk_b) + 
					sum(k.scrake) + sum(k.fp) + sum(k.qp) + sum(k.boss) + sum(k.custom) as total_kills,

					session.id as max_damage_session_id,
					sum(damage_dealt) as max_damage
				FROM session
				INNER JOIN wave_stats ws ON ws.session_id = session.id
				INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
				INNER JOIN wave_stats_player_kills k ON k.player_stats_id = wsp.id
				WHERE session.id = session_id
				GROUP BY session.id, wsp.player_id, wsp.perk
			) as new
			ON DUPLICATE KEY UPDATE
				user_weekly_stats_perk.total_games = user_weekly_stats_perk.total_games + new.total_games,
				user_weekly_stats_perk.total_waves = user_weekly_stats_perk.total_waves + new.total_waves,
				user_weekly_stats_perk.playtime_seconds = user_weekly_stats_perk.playtime_seconds + new.playtime_seconds,
				user_weekly_stats_perk.deaths = user_weekly_stats_perk.deaths + new.deaths,

				user_weekly_stats_perk.shots_fired = user_weekly_stats_perk.shots_fired + new.shots_fired,
				user_weekly_stats_perk.shots_hit = user_weekly_stats_perk.shots_hit + new.shots_hit,
				user_weekly_stats_perk.shots_hs = user_weekly_stats_perk.shots_hs + new.shots_hs,

				user_weekly_stats_perk.dosh_earned = user_weekly_stats_perk.dosh_earned + new.dosh_earned,
				user_weekly_stats_perk.heals_given = user_weekly_stats_perk.heals_given + new.heals_given,
				user_weekly_stats_perk.heals_recv = user_weekly_stats_perk.heals_recv + new.heals_recv,

				user_weekly_stats_perk.damage_dealt = user_weekly_stats_perk.damage_dealt + new.damage_dealt,
				user_weekly_stats_perk.damage_taken = user_weekly_stats_perk.damage_taken + new.damage_taken,
				
				user_weekly_stats_perk.large_kills = user_weekly_stats_perk.large_kills + new.large_kills,
				user_weekly_stats_perk.total_kills = user_weekly_stats_perk.total_kills + new.total_kills,

				user_weekly_stats_perk.zedtime_count = user_weekly_stats_perk.zedtime_count + new.zedtime_count,
				user_weekly_stats_perk.zedtime_length = user_weekly_stats_perk.zedtime_length + new.zedtime_length,

				user_weekly_stats_perk.buffs_active_length = user_weekly_stats_perk.buffs_active_length + new.buffs_active_length,
				user_weekly_stats_perk.buffs_total_length = user_weekly_stats_perk.buffs_total_length + new.buffs_total_length,

				user_weekly_stats_perk.large_kills = user_weekly_stats_perk.large_kills + new.large_kills,
				user_weekly_stats_perk.total_kills = user_weekly_stats_perk.total_kills + new.total_kills,

				max_damage_session_id = new.max_damage_session_id,
				max_damage = new.max_damage;
		END;
	`)
	tx.Exec(`
		DROP PROCEDURE IF EXISTS fill_weekly_user_stats;
		CREATE PROCEDURE fill_weekly_user_stats()
		BEGIN
			DECLARE session_id INT DEFAULT NULL;
			DECLARE done TINYINT DEFAULT FALSE;

			DECLARE sessions_cursor 
				CURSOR FOR SELECT id FROM session WHERE is_completed;
				
			DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;

			START TRANSACTION;

			OPEN sessions_cursor;

			sessions_loop:
			LOOP
				FETCH NEXT FROM sessions_cursor INTO session_id;

				IF done THEN
					LEAVE sessions_loop; 
				ELSE
					CALL update_user_stats_weekly(session_id);
				END IF;
			END LOOP;

			CLOSE sessions_cursor;

			COMMIT;
		END;
	`)

	return err
}
