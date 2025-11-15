package analytics

import (
	"database/sql"
	"fmt"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

func ExecuteHistoricalQuery(
	db *sql.DB, query string, args ...any,
) ([]*models.PeriodData, error) {
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

	rows, err := db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	items := []*models.PeriodData{}
	for rows.Next() {
		item := models.PeriodData{}

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
