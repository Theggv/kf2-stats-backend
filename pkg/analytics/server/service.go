package server

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/theggv/kf2-stats-backend/pkg/analytics"
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
	case analytics.Hour:
		period = "HOUR(session.started_at)"
	case analytics.Day, analytics.Week:
		period = "DAY(session.started_at)"
	case analytics.Month:
		period = "MONTH(session.started_at)"
	case analytics.Year:
		period = "YEAR(session.started_at)"
	default:
		return nil, analytics.NewIncorrectPeriod(req.Period)
	}

	conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	sql := fmt.Sprintf(`
		SELECT 
			count(*) AS times_played, 
			%v as period
		FROM session
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
	case analytics.Day, analytics.Week:
		period = "DAY(session.started_at)"
	case analytics.Month:
		period = "MONTH(session.started_at)"
	case analytics.Year:
		period = "YEAR(session.started_at)"
	default:
		return nil, analytics.NewIncorrectPeriod(req.Period)
	}

	conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?", "session.completed_at IS NOT NULL")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	sql := fmt.Sprintf(`
		SELECT 
			sum(timestampdiff(MINUTE, started_at, completed_at)), 
			%v as period
		FROM session
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
	case analytics.Hour:
		period = "HOUR(session.started_at)"
	case analytics.Day, analytics.Week:
		period = "DAY(session.started_at)"
	case analytics.Month:
		period = "MONTH(session.started_at)"
	case analytics.Year:
		period = "YEAR(session.started_at)"
	default:
		return nil, analytics.NewIncorrectPeriod(req.Period)
	}

	conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
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
