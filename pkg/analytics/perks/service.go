package perks

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type PerksAnalyticsService struct {
	db *sql.DB
}

func NewPerksAnalyticsService(db *sql.DB) *PerksAnalyticsService {
	service := PerksAnalyticsService{
		db: db,
	}

	return &service
}

func (s *PerksAnalyticsService) GetPerksPlayTime(
	req PerksPlayTimeRequest,
) (*[]PerkStats, error) {
	conds := make([]string, 0)
	args := make([]interface{}, 0)

	if req.ServerId != 0 {
		conds = append(conds, "session.server_id = ?")
		args = append(args, req.ServerId)
	}

	if req.UserId != 0 {
		conds = append(conds, "wsp.player_id = ?")
		args = append(args, req.UserId)
	}

	conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	sql := fmt.Sprintf(`
		SELECT 
			floor(sum(timestampdiff(SECOND, ws.started_at, ws.completed_at)) / 3600) as time_played_hours,
			wsp.perk
		FROM session
		INNER JOIN wave_stats ws ON session.id = ws.session_id
		INNER JOIN wave_stats_player wsp ON ws.id = wsp.stats_id
		WHERE %v
		GROUP BY wsp.perk
		HAVING time_played_hours > 0
		ORDER BY wsp.perk`,
		strings.Join(conds, " AND "),
	)

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}

	items := []PerkStats{}
	for rows.Next() {
		item := PerkStats{}

		err = rows.Scan(&item.Count, &item.Perk)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return &items, nil
}

func (s *PerksAnalyticsService) GetPerksKills(
	req PerksKillsRequest,
) (*[]PerkStats, error) {
	conds := make([]string, 0)
	args := make([]interface{}, 0)

	if req.ServerId != 0 {
		conds = append(conds, "session.server_id = ?")
		args = append(args, req.ServerId)
	}

	if req.UserId != 0 {
		conds = append(conds, "wsp.player_id = ?")
		args = append(args, req.UserId)
	}

	conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	sql := fmt.Sprintf(`
		SELECT
			sum(kills.total) as kills,
			wsp.perk
		FROM session
		INNER JOIN wave_stats ws ON session.id = ws.session_id
		INNER JOIN wave_stats_player wsp ON ws.id = wsp.stats_id
		INNER JOIN aggregated_kills kills ON wsp.id = kills.player_stats_id
		WHERE %v
		GROUP BY wsp.perk
		HAVING kills > 0
		ORDER BY wsp.perk`,
		strings.Join(conds, " AND "),
	)

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}

	items := []PerkStats{}
	for rows.Next() {
		item := PerkStats{}

		err = rows.Scan(&item.Count, &item.Perk)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return &items, nil
}

func newIncorrectPeriod(period int) error {
	return errors.New(fmt.Sprintf("expected TimePeriod enum, got %v", period))
}
