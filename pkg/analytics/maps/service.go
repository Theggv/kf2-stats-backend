package maps

import (
	"database/sql"
	"fmt"
	"strings"
)

type MapAnalyticsService struct {
	db *sql.DB
}

func NewMapAnalyticsService(db *sql.DB) *MapAnalyticsService {
	service := MapAnalyticsService{
		db: db,
	}

	return &service
}

func (s *MapAnalyticsService) GetMapAnalytics(
	req MapAnalyticsRequest,
) ([]*MapAnalytics, error) {
	limit := clampLimit(req.Limit)

	conds := make([]string, 0)
	args := make([]any, 0)

	conds = append(conds, "session.started_at is not null")

	if req.ServerId != 0 {
		conds = append(conds, "session.server_id = ?")
		args = append(args, req.ServerId)
	}

	conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	stmt := fmt.Sprintf(`
		WITH maps_rating AS (
			SELECT
				maps.id as map_id,
				maps.name as map_name,
				count(session.id) as value
			FROM session
			INNER JOIN maps ON maps.id = session.map_id
			WHERE %v
			GROUP BY session.map_id
			ORDER BY value DESC
			LIMIT %v
		), other AS (
			SELECT count(*) as value
			FROM session
			LEFT JOIN rating ON rating.map_id = session.map_id
			WHERE rating.map_id is null AND %v
		), result AS (
			SELECT
				maps.id as map_id,
				maps.name as map_name,
				cte.value as value
			FROM rating cte
			INNER JOIN maps ON maps.id = cte.map_id

			UNION ALL

			SELECT
				0 as map_id,
				"Other" as map_name,
				value
			FROM other
			WHERE value > 0
		)
		SELECT map_id, map_name, value 
		FROM result
		`,
		strings.Join(conds, " AND "), limit, strings.Join(conds, " AND "),
	)

	rows, err := s.db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	items := []*MapAnalytics{}
	for rows.Next() {
		item := MapAnalytics{}

		err = rows.Scan(&item.MapId, &item.MapName, &item.Count)
		if err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	return items, nil
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return 10
	} else if limit > 100 {
		return 100
	}
	return limit
}
