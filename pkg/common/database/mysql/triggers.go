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

	return err
}
