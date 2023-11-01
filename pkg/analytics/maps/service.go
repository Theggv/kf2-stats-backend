package maps

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/theggv/kf2-stats-backend/pkg/common/util"
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
) (*[]MapAnalytics, error) {
	limit := clampLimit(req.Limit)

	conds := make([]string, 0)
	args := make([]interface{}, 0)

	conds = append(conds, "session.status in (2,3,4)")

	if req.ServerId != 0 {
		conds = append(conds, "session.server_id = ?")
		args = append(args, req.ServerId)
	}

	conds = append(conds, "substr(session.completed_at, 1, 10) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	sql := fmt.Sprintf(`
		SELECT
			maps.id,
			maps.name,
			count(session.id) AS times_played
		FROM session
		INNER JOIN maps ON maps.id = session.map_id
		WHERE %v
		GROUP BY session.map_id
		ORDER BY times_played DESC
		LIMIT %v`, strings.Join(conds, " AND "), limit,
	)

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}

	items := []MapAnalytics{}
	for rows.Next() {
		item := MapAnalytics{}

		err = rows.Scan(&item.MapId, &item.MapName, &item.Count)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	mapIds := []int{}
	for _, item := range items {
		mapIds = append(mapIds, item.MapId)
	}

	if len(mapIds) > 0 {
		sql := fmt.Sprintf(`
			SELECT count(session.id) AS times_played
			FROM session
			INNER JOIN maps ON maps.id = session.map_id
			WHERE %v AND session.map_id not in (%v)`,
			strings.Join(conds, " AND "),
			util.IntArrayToString(mapIds, ","),
		)

		stmt, err := s.db.Prepare(sql)
		if err != nil {
			return nil, err
		}

		rows, err := stmt.Query(args...)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			other := MapAnalytics{
				MapId:   0,
				MapName: "Other",
			}

			err = rows.Scan(&other.Count)
			if err != nil {
				return nil, err
			}

			if other.Count > 0 {
				items = append(items, other)
			}
		}
	}

	return &items, nil
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return 10
	} else if limit > 100 {
		return 100
	}
	return limit
}
