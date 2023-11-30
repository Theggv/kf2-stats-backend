package perks

import (
	"database/sql"
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
		conds = append(conds, "aggr.user_id = ?")
		args = append(args, req.UserId)
	}

	conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	sql := fmt.Sprintf(`
		SELECT 
			floor(sum(aggr.playtime_seconds) / 3600) as time_played_hours,
			aggr.perk
		FROM session
		INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
		WHERE %v
		GROUP BY aggr.perk
		HAVING time_played_hours > 0
		ORDER BY aggr.perk`,
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
		conds = append(conds, "aggr.user_id = ?")
		args = append(args, req.UserId)
	}

	conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	sql := fmt.Sprintf(`
		SELECT
			sum(kills.total) as kills,
			aggr.perk
		FROM session
		INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
		INNER JOIN session_aggregated_kills kills ON kills.id = aggr.id
		WHERE %v
		GROUP BY aggr.perk
		HAVING kills > 0
		ORDER BY aggr.perk`,
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
