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

	conds = append(conds, "session.is_completed = TRUE")

	if req.ServerId != 0 {
		conds = append(conds, "session.server_id = ?")
		args = append(args, req.ServerId)
	}

	var period string
	switch req.Period {
	case analytics.Hour:
		period = "HOUR(session.started_at)"
	case analytics.Day, analytics.Week:
		period = "DAY(session.started_at)"
	case analytics.Month:
		period = "DATE_FORMAT(session.started_at, '%Y-%m-01 00:00:00')"
	case analytics.Year:
		period = "DATE_FORMAT(session.started_at, '%Y-01-01 00:00:00')"
	case analytics.Date:
		period = "DATE_FORMAT(session.started_at, '%Y-%m-%d 00:00:00')"
	case analytics.DateHour:
		period = "DATE_FORMAT(session.started_at, '%Y-%m-%d %H:00:00')"
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
		period = "DATE_FORMAT(session.started_at, '%Y-%m-01 00:00:00')"
	case analytics.Year:
		period = "DATE_FORMAT(session.started_at, '%Y-01-01 00:00:00')"
	case analytics.Date:
		period = "DATE_FORMAT(session.started_at, '%Y-%m-%d 00:00:00')"
	case analytics.DateHour:
		period = "DATE_FORMAT(session.started_at, '%Y-%m-%d %H:00:00')"
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

	conds = append(conds, "session.is_completed = TRUE")

	if req.ServerId != 0 {
		conds = append(conds, "session.server_id = ?")
		args = append(args, req.ServerId)
	}

	var period string
	switch req.Period {
	case analytics.Hour:
		period = "HOUR(session.started_at)"
	case analytics.Day, analytics.Week:
		period = "DAY(session.started_at)"
	case analytics.Month:
		period = "DATE_FORMAT(session.started_at, '%Y-%m-01 00:00:00')"
	case analytics.Year:
		period = "DATE_FORMAT(session.started_at, '%Y-01-01 00:00:00')"
	case analytics.Date:
		period = "DATE_FORMAT(session.started_at, '%Y-%m-%d 00:00:00')"
	case analytics.DateHour:
		period = "DATE_FORMAT(session.started_at, '%Y-%m-%d %H:00:00')"
	default:
		return nil, analytics.NewIncorrectPeriod(req.Period)
	}

	conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	sql := fmt.Sprintf(`
		SELECT
			count(DISTINCT aggr.user_id), 
			%v as period
		FROM session
		INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
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

func (s *ServerAnalyticsService) GetPopularServers() (*PopularServersResponse, error) {
	sql := fmt.Sprintf(`
		SELECT
			server.id,
			server.name,
			total_sessions,
			total_users,
			diff
		FROM (
			SELECT
				session.server_id as server_id,
				count(DISTINCT session.id) as total_sessions,
				count(DISTINCT aggr.user_id) as total_users,
				min(session.diff) as diff
			FROM session
			INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
			WHERE DATE(session.started_at) BETWEEN now() - interval 30 day AND now()
			GROUP BY server_id
			ORDER BY total_users desc
			LIMIT 5
		) t
		INNER JOIN server ON server.id = t.server_id`,
	)

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	items := []PopularServersResponseItem{}
	for rows.Next() {
		item := PopularServersResponseItem{}

		err = rows.Scan(&item.Id, &item.Name,
			&item.TotalSessions, &item.TotalUsers, &item.Difficulty,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return &PopularServersResponse{
		Items: items,
	}, nil
}

func (s *ServerAnalyticsService) GetCurrentOnline() (*TotalOnlineResponse, error) {
	sql := fmt.Sprintf(`
		SELECT count(*) as total_online
		FROM users_activity
		WHERE current_session_id is not null`,
	)

	res := TotalOnlineResponse{}
	err := s.db.QueryRow(sql).Scan(&res.Count)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
