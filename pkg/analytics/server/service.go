package server

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type ServerAnalyticsService struct {
	db *sql.DB
}

func NewServerAnalyticsService(db *sql.DB) *ServerAnalyticsService {
	service := ServerAnalyticsService{
		db: db,
	}

	return &service
}

func (s *ServerAnalyticsService) GetSessionCount(
	req SessionCountRequest,
) (*[]PeriodData, error) {
	conds := make([]string, 0)
	args := make([]interface{}, 0)

	conds = append(conds, "session.is_completed = TRUE", "session.server_id = ?")
	args = append(args, req.ServerId)

	var period string
	switch req.Period {
	case Hour:
		period = "HOUR(session.completed_at)"
	case Day, Week:
		period = "DAY(session.completed_at)"
	case Month:
		period = "MONTH(session.completed_at)"
	case Year:
		period = "YEAR(session.completed_at)"
	default:
		return nil, newIncorrectPeriod(req.Period)
	}

	conds = append(conds, "DATE(session.completed_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	sql := fmt.Sprintf(`
		SELECT 
			count(*) AS times_played, 
			%v as period
		FROM session
		INNER JOIN server ON server.id = session.server_id
		WHERE %v
		GROUP BY period
		ORDER BY period`,
		period, strings.Join(conds, " AND "),
	)

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}

	items := []PeriodData{}
	for rows.Next() {
		item := PeriodData{}

		err = rows.Scan(&item.Count, &item.Period)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return &items, nil
}

func (s *ServerAnalyticsService) GetUsageInMinutes(
	req UsageInMinutesRequest,
) (*[]PeriodData, error) {
	conds := make([]string, 0)
	args := make([]interface{}, 0)

	conds = append(conds, "session.is_completed = TRUE", "session.server_id = ?")
	args = append(args, req.ServerId)

	var period string
	switch req.Period {
	case Day, Week:
		period = "DAY(session.completed_at)"
	case Month:
		period = "MONTH(session.completed_at)"
	case Year:
		period = "YEAR(session.completed_at)"
	default:
		return nil, newIncorrectPeriod(req.Period)
	}

	conds = append(conds, "DATE(session.completed_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	sql := fmt.Sprintf(`
		SELECT 
			sum(timestampdiff(MINUTE, started_at, completed_at)), 
			%v as period
		FROM session
		INNER JOIN server ON server.id = session.server_id
		WHERE %v
		GROUP BY period
		ORDER BY period`,
		period, strings.Join(conds, " AND "),
	)

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}

	items := []PeriodData{}
	for rows.Next() {
		item := PeriodData{}

		err = rows.Scan(&item.Count, &item.Period)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return &items, nil
}

func (s *ServerAnalyticsService) GetPlayersOnline(
	req PlayersOnlineRequest,
) (*[]PeriodData, error) {
	conds := make([]string, 0)
	args := make([]interface{}, 0)

	conds = append(conds, "session.is_completed = TRUE", "session.server_id = ?")
	args = append(args, req.ServerId)

	var period string
	switch req.Period {
	case Hour:
		period = "HOUR(session.completed_at)"
	case Day, Week:
		period = "DAY(session.completed_at)"
	case Month:
		period = "MONTH(session.completed_at)"
	case Year:
		period = "YEAR(session.completed_at)"
	default:
		return nil, newIncorrectPeriod(req.Period)
	}

	conds = append(conds, "DATE(session.completed_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	sql := fmt.Sprintf(`
		SELECT
			count(DISTINCT wsp.player_id), 
			%v as period
		FROM session
		INNER JOIN wave_stats ws on ws.session_id = session.id
		INNER JOIN wave_stats_player wsp on wsp.stats_id = ws.id
		WHERE %v
		GROUP BY period
		ORDER BY period`,
		period, strings.Join(conds, " AND "),
	)

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}

	items := []PeriodData{}
	for rows.Next() {
		item := PeriodData{}

		err = rows.Scan(&item.Count, &item.Period)
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
