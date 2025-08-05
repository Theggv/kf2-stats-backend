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
) ([]*PeriodData, error) {
	conds := make([]string, 0)
	args := make([]any, 0)

	conds = append(conds, "session.started_at is not null")

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

	if req.From != nil && req.To != nil {
		conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	}

	stmt := fmt.Sprintf(`
		SELECT
			%v as period,
			count(*) as value
		from session
		WHERE %v
		GROUP BY period
		ORDER BY period`,
		period, strings.Join(conds, " AND "),
	)

	return s.executeHistoricalQuery(stmt, args...)
}

func (s *ServerAnalyticsService) GetUsageInMinutes(
	req UsageInMinutesRequest,
) ([]*PeriodData, error) {
	conds := make([]string, 0)
	args := make([]any, 0)

	conds = append(conds, "session.started_at is not null", "session.completed_at is not null", "session.server_id = ?")
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

	if req.From != nil && req.To != nil {
		conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	}

	stmt := fmt.Sprintf(`
		SELECT
			%v as period,
			sum(timestampdiff(MINUTE, started_at, completed_at)) as value
		from session
		WHERE %v
		GROUP BY period
		ORDER BY period`,
		period, strings.Join(conds, " AND "),
	)

	return s.executeHistoricalQuery(stmt, args...)
}

func (s *ServerAnalyticsService) GetPlayersOnline(
	req PlayersOnlineRequest,
) ([]*PeriodData, error) {
	conds := make([]string, 0)
	args := make([]any, 0)

	conds = append(conds, "session.started_at is not null")

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

	if req.From != nil && req.To != nil {
		conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	}

	stmt := fmt.Sprintf(`
		SELECT DISTINCT
			period,
			count(user_id) over (partition by period) as value
		FROM (
			SELECT DISTINCT
				%v as period,
				aggr.user_id as user_id
			FROM session
			INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
			WHERE %v
		) t`,
		period, strings.Join(conds, " AND "),
	)

	return s.executeHistoricalQuery(stmt, args...)
}

func (s *ServerAnalyticsService) GetPopularServers() (*PopularServersResponse, error) {
	stmt := `
		WITH server_rating AS (
			SELECT
				session.server_id as server_id,
				count(DISTINCT session.id) as total_sessions,
				count(DISTINCT aggr.user_id) as total_users,
				min(session.diff) as diff
			FROM session
			INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
			GROUP BY server_id
			ORDER BY total_users desc
			LIMIT 5
		)
		SELECT 
			server.id as server_id,
			server.name as server_name,
			cte.total_sessions as total_session,
			cte.total_users as total_users,
			cte.diff as difficulty
		FROM server_rating cte
		INNER JOIN server ON server.id = cte.server_id`

	rows, err := s.db.Query(stmt)
	if err != nil {
		return nil, err
	}

	items := []*PopularServersResponseItem{}
	for rows.Next() {
		item := PopularServersResponseItem{}

		err = rows.Scan(&item.Id, &item.Name,
			&item.TotalSessions, &item.TotalUsers, &item.Difficulty,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	return &PopularServersResponse{
		Items: items,
	}, nil
}

func (s *ServerAnalyticsService) GetCurrentOnline() (*TotalOnlineResponse, error) {
	stmt := `
		SELECT count(*) as total_online
		FROM users_activity
		WHERE updated_at >= now() - interval 1 minute`

	res := TotalOnlineResponse{}
	err := s.db.QueryRow(stmt).Scan(&res.Count)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *ServerAnalyticsService) executeHistoricalQuery(query string, args ...any) ([]*PeriodData, error) {
	stmt := fmt.Sprintf(`
		WITH historical_data AS (
			%v
		), with_lag AS (
			SELECT
				cte.*,
				LAG(value, 1, 0) OVER w AS prev
			FROM historical_data cte
			WINDOW w AS (ORDER BY period)
		)
		SELECT 
			period,
			value,
			prev,
			value - prev AS diff,
			first_value(value) OVER value_frame AS max_value,
			avg(value) OVER trend_frame AS trend_value
		FROM with_lag
		WINDOW 
			value_frame AS (ORDER BY value DESC),
			trend_frame AS (ORDER BY period ROWS BETWEEN 5 PRECEDING AND CURRENT ROW)
		`, query,
	)

	rows, err := s.db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	items := []*PeriodData{}
	for rows.Next() {
		item := PeriodData{}

		err = rows.Scan(
			&item.Period, &item.Value, &item.PreviousValue,
			&item.Difference, &item.MaxValue, &item.Trend,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	return items, nil
}
