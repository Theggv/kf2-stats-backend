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
				session.id <> 0 AND session.id NOT IN (tbl.max_id) AND 
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
				session.id <> 0 AND 
				session.status IN (0, 1) AND 
				timestampdiff(MINUTE, gd.updated_at, CURRENT_TIMESTAMP) > minutes;
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
			DO SLEEP(3);
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
		END;
	`)

	return err
}
