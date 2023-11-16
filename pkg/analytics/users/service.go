package users

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

type UserAnalyticsService struct {
	db *sql.DB
}

func NewUserAnalyticsService(db *sql.DB) *UserAnalyticsService {
	service := UserAnalyticsService{
		db: db,
	}

	return &service
}

func (s *UserAnalyticsService) GetUserAnalytics(
	req UserAnalyticsRequest,
) (*UserAnalyticsResponse, error) {
	res := UserAnalyticsResponse{}

	conds := []string{
		"wsp.player_id = ?",
		"DATE(session.started_at) BETWEEN ? AND ?",
	}
	args := []interface{}{
		req.UserId, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"),
	}

	sql := fmt.Sprintf(`
		SELECT
			count(distinct session.id) as total_games,
			coalesce(sum(kills.total), 0) as total_kills,
			coalesce(sum(wsp.is_dead = 1), 0) as total_deaths
		FROM session
		INNER JOIN wave_stats ws ON ws.session_id = session.id
		INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
		INNER JOIN aggregated_kills kills ON wsp.id = kills.player_stats_id
		WHERE %v`, strings.Join(conds, " AND "),
	)

	err := s.db.QueryRow(sql, args...).Scan(&res.Games, &res.Kills, &res.Deaths)
	if err != nil {
		return nil, err
	}

	sql = fmt.Sprintf(`
		SELECT floor(coalesce(sum(t.seconds), 0) / 60) as total_minutes
		FROM (
			SELECT timestampdiff(SECOND, ws.started_at, ws.completed_at) as seconds
			FROM session
			INNER JOIN wave_stats ws ON ws.session_id = session.id
			INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
			WHERE %v
			GROUP BY ws.id
		) t`, strings.Join(conds, " AND "),
	)

	err = s.db.QueryRow(sql, args...).Scan(&res.Minutes)
	if err != nil {
		return nil, err
	}

	sql = fmt.Sprintf(`
		SELECT count(t.status) as total_wins
		FROM (
			SELECT status
			FROM session
			INNER JOIN wave_stats ws ON ws.session_id = session.id
			INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
			WHERE %v
			GROUP BY session.id
		) t
		WHERE status = %v`, strings.Join(conds, " AND "), models.Win,
	)

	err = s.db.QueryRow(sql, args...).Scan(&res.Wins)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return 10
	} else if limit > 100 {
		return 100
	}
	return limit
}
