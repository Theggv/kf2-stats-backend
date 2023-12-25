package leaderboards

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
	"github.com/theggv/kf2-stats-backend/pkg/users"
)

type LeaderBoardsService struct {
	db *sql.DB

	userService *users.UserService
}

func NewLeaderBoardsService(db *sql.DB) *LeaderBoardsService {
	service := LeaderBoardsService{
		db: db,
	}

	return &service
}

func (s *LeaderBoardsService) Inject(
	userService *users.UserService,
) {
	s.userService = userService
}

func (s *LeaderBoardsService) getLeaderBoard(
	req LeaderBoardsRequest,
) (*LeaderBoardsResponse, error) {

	var (
		userIds []int
		err     error
	)
	switch req.Type {
	case AverageZedtime:
		userIds, err = s.getZedtimeTop(req)
	case HsAccuracy:
		userIds, err = s.getHSAccuracyTop(req)
	case MostDamage:
		userIds, err = s.getMostDamageTop(req)
	default:
		userIds, err = s.getLeaderboardIds(req)
	}

	if err != nil {
		return nil, err
	}

	conds := make([]string, 0)
	args := make([]interface{}, 0)

	// field args
	args = append(args,
		req.Perk, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"),
		req.Perk, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"),
		req.From.Format("2006-01-02"), req.To.Format("2006-01-02"),
	)

	// where args
	conds = append(conds, fmt.Sprintf("user_id IN (%v)", util.IntArrayToString(userIds, ",")))

	conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	if req.Perk != 0 {
		conds = append(conds, "perk = ?")
		args = append(args, req.Perk)
	}

	sql := fmt.Sprintf(`
		SELECT
			users.id,
			users.name,
			users.auth_type,
			users.auth_id,
			t.total_games,
			t.total_deaths,
			t.total_damage,
			t.total_kills,
			t.total_large_kills,
			t.total_heals,
			t.total_playtime,
			get_avg_hs_acc(users.id, ?, ?, ?) as avg_hs_acc,
			get_avg_acc(users.id, ?, ?, ?) as avg_acc,
			get_avg_zt(users.id, ?, ?) as avg_zt
		FROM (
			SELECT
				user_id,
				count(distinct session.id) as total_games,
				coalesce(sum(deaths), 0) as total_deaths,
				coalesce(sum(damage_dealt), 0) as total_damage,
				coalesce(sum(kills.total), 0) as total_kills,
				coalesce(sum(kills.large), 0) as total_large_kills,
				coalesce(sum(heals_given), 0) as total_heals,
				floor(coalesce(sum(playtime_seconds), 0) / 3600) as total_playtime
			FROM session
			INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
			INNER JOIN session_aggregated_kills kills ON aggr.id = kills.id
			WHERE %v
			GROUP BY user_id
		) t
		INNER JOIN users ON users.id = t.user_id
		GROUP BY user_id
		ORDER BY FIELD(users.id, %v)
		`, strings.Join(conds, " AND "), util.IntArrayToString(userIds, ","),
	)

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	res := LeaderBoardsResponse{
		Items: []*LeaderBoardsResponseItem{},
	}
	steamIdSet := make(map[string]bool)

	for rows.Next() {
		item := LeaderBoardsResponseItem{}

		err := rows.Scan(
			&item.Id, &item.Name,
			&item.Type, &item.AuthId,
			&item.TotalGames, &item.TotalDeaths,
			&item.TotalDamage, &item.TotalKills,
			&item.TotalLargeKills, &item.TotalHeals,
			&item.TotalPlaytime, &item.HSAccuracy,
			&item.Accuracy, &item.AverageZedtime,
		)
		if err != nil {
			return nil, err
		}

		if item.Type == models.Steam {
			steamIdSet[item.AuthId] = true
		}

		res.Items = append(res.Items, &item)
	}

	// Join most damage games
	{
		conds := make([]string, 0)
		args := make([]interface{}, 0)

		conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

		if req.Perk != 0 {
			conds = append(conds, "perk = ?")
			args = append(args, req.Perk)
		}

		conds = append(conds, fmt.Sprintf("user_id IN (%v)", util.IntArrayToString(userIds, ",")))

		sql := fmt.Sprintf(`
			SELECT
				user_id,
				ANY_VALUE(session_id) as session_id,
				max(metric) as metric
			FROM (
				SELECT
					user_id,
					session_id,
					sum(aggr.damage_dealt) as metric
				FROM session
				INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
				WHERE %v
				GROUP BY session_id, user_id
				ORDER BY metric DESC
			) t
			GROUP BY user_id`, strings.Join(conds, " AND "),
		)

		rows, err := s.db.Query(sql, args...)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var userId int
			var match MostDamageMatch

			err := rows.Scan(&userId, &match.SessionId, &match.Value)
			if err != nil {
				return nil, err
			}

			for _, item := range res.Items {
				if item.Id != userId {
					continue
				}

				item.MostDamage = &match
			}
		}
	}

	// Join steam data
	{
		steamIds := []string{}
		for key := range steamIdSet {
			steamIds = append(steamIds, key)
		}

		steamData, err := s.userService.GetSteamData(steamIds)
		if err != nil {
			return nil, err
		}

		for _, item := range res.Items {
			if data, ok := steamData[item.AuthId]; ok {
				item.Avatar = &data.Avatar
				item.ProfileUrl = &data.ProfileUrl
			}
		}
	}
	return &res, nil
}

func (s *LeaderBoardsService) getHSAccuracyTop(
	req LeaderBoardsRequest,
) ([]int, error) {
	conds := make([]string, 0)
	args := make([]interface{}, 0)

	if req.Perk == 0 {
		return nil, errors.New(fmt.Sprintf("perk cannot be null with selected type"))
	}

	args = append(args, req.Perk, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	conds = append(conds, "perk = ?")
	args = append(args, req.Perk)

	conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	sql := fmt.Sprintf(`
		SELECT
			user_id,
			get_avg_hs_acc(user_id, ?, ?, ?) as metric
		FROM (
			SELECT
				user_id,
				count(distinct session.id) as total_games
			FROM session
			INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
			WHERE %v
			GROUP BY user_id
			HAVING total_games >= 10
		) t
		ORDER BY metric DESC
		LIMIT 50`, strings.Join(conds, " AND "),
	)

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	items := []int{}
	for rows.Next() {
		var item int
		var unused float64
		err := rows.Scan(&item, &unused)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func (s *LeaderBoardsService) getZedtimeTop(
	req LeaderBoardsRequest,
) ([]int, error) {
	conds := make([]string, 0)
	args := make([]interface{}, 0)

	conds = append(conds, fmt.Sprintf("perk = %v", models.Commando))
	conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	sql := fmt.Sprintf(`
		SELECT
			user_id,
			get_avg_zt(user_id, ?, ?) as metric
		FROM (
			SELECT
				user_id,
				count(distinct session.id) as total_games
			FROM session
			INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
			WHERE %v
			GROUP BY user_id
			HAVING total_games >= 10
		) t
		ORDER BY metric DESC
		LIMIT 50`, strings.Join(conds, " AND "),
	)

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	items := []int{}
	for rows.Next() {
		var item int
		var unused float64
		err := rows.Scan(&item, &unused)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func (s *LeaderBoardsService) getMostDamageTop(
	req LeaderBoardsRequest,
) ([]int, error) {
	conds := make([]string, 0)
	args := make([]interface{}, 0)

	conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	if req.Perk != 0 {
		conds = append(conds, "perk = ?")
		args = append(args, req.Perk)
	}

	sql := fmt.Sprintf(`
		SELECT
			user_id,
			max(metric) as metric
		FROM (
			SELECT
				user_id,
				session_id,
				sum(aggr.damage_dealt) as metric
			FROM session
			INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
			WHERE %v
			GROUP BY session_id, user_id
			ORDER BY metric DESC
		) t
		GROUP BY user_id
		ORDER BY metric desc
		LIMIT 50`, strings.Join(conds, " AND "),
	)

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	items := []int{}
	for rows.Next() {
		var item int
		var unused float64
		err := rows.Scan(&item, &unused)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func (s *LeaderBoardsService) getLeaderboardIds(
	req LeaderBoardsRequest,
) ([]int, error) {
	conds := make([]string, 0)
	args := make([]interface{}, 0)

	var metric string
	switch req.Type {
	case TotalGames:
		metric = "count(distinct session.id) as metric"
	case TotalDeaths:
		metric = "coalesce(sum(deaths), 0) as metric"
	case TotalDamage:
		metric = "coalesce(sum(damage_dealt), 0) as metric"
	case TotalKills:
		metric = "coalesce(sum(kills.total), 0) as metric"
	case TotalLargeKills:
		metric = "coalesce(sum(kills.large), 0) as metric"
	case TotalHeals:
		metric = "coalesce(sum(heals_given), 0) as metric"
	case TotalPlaytime:
		metric = "floor(coalesce(sum(playtime_seconds), 0) / 3600) as metric"
	}

	conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	if req.Perk != 0 {
		conds = append(conds, "perk = ?")
		args = append(args, req.Perk)
	}

	sql := fmt.Sprintf(`
		SELECT
			users.id,
			t.metric
		FROM (
			SELECT
				user_id,
				count(distinct session.id) as total_games,
				%v
			FROM session
			INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
			INNER JOIN session_aggregated_kills kills ON aggr.id = kills.id
			WHERE %v
			GROUP BY user_id
			HAVING total_games >= 10
			ORDER BY metric DESC
			LIMIT 50
		) t
		INNER JOIN users ON users.id = t.user_id
		GROUP BY user_id
		ORDER BY metric DESC, name ASC
		`, metric, strings.Join(conds, " AND "),
	)

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	items := []int{}
	for rows.Next() {
		var item int
		var unused float64
		err := rows.Scan(&item, &unused)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}
