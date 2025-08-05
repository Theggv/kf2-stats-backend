package leaderboards

import (
	"database/sql"
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

type userIdResponse struct {
	Ids   []int
	Total int
}

func (s *LeaderBoardsService) getLeaderBoard(
	req LeaderBoardsRequest,
) (*LeaderBoardsResponse, error) {
	var (
		userData *userIdResponse
		err      error
	)

	if req.Page < 0 {
		req.Page = 0
	}

	err = s.validateRequest(req)
	if err != nil {
		return nil, err
	}

	switch req.OrderBy {
	case MostDamage:
		userData, err = s.getMostDamageTop(req)
	default:
		userData, err = s.getLeaderboardIds(req)
	}

	if err != nil {
		return nil, err
	}

	if len(userData.Ids) == 0 {
		return &LeaderBoardsResponse{
			Items: []*LeaderBoardsResponseItem{},
			Metadata: &models.PaginationResponse{
				Page:           req.Page,
				ResultsPerPage: 50,
				TotalResults:   0,
			},
		}, nil
	}

	fields := []string{
		"users.id as user_id",
		"users.name as user_name",
		"users.auth_type as auth_type",
		"users.auth_id as auth_id",
		"t.total_games as total_games",
		"t.total_deaths as total_deaths",
		"t.total_damage as total_damage",
		"t.total_kills as total_kills",
		"t.large_kills as large_kills",
		"t.total_heals as total_heals",
		"t.total_playtime as total_playtime",
	}

	conds := make([]string, 0)
	args := make([]any, 0)

	tableName := "user_weekly_stats_total"
	if req.Perk != 0 {
		tableName = "user_weekly_stats_perk"

		fields = append(fields,
			"greatest(0, least(coalesce(sum(shots_hs) / sum(shots_hit), 0), 1)) as avg_hs_acc",
			"greatest(0, least(coalesce(sum(shots_hit) / sum(shots_fired), 0), 1)) as avg_acc",
			"coalesce(sum(zedtime_length) / sum(zedtime_count), 0) as avg_zt",
			"coalesce(sum(buffs_active_length) / sum(buffs_total_length), 0) as avg_buffs_uptime",
		)
	}

	// prepare subquery
	var sq string
	{
		fields := []string{
			"coalesce(sum(total_games), 0) as total_games",
			"coalesce(sum(deaths), 0) as total_deaths",
			"coalesce(sum(damage_dealt), 0) as total_damage",
			"coalesce(sum(heals_given), 0) as total_heals",
			"coalesce(sum(total_kills), 0) as total_kills",
			"coalesce(sum(large_kills), 0) as large_kills",
			"coalesce(sum(shots_hs), 0) as shots_hs",
			"coalesce(sum(shots_hit), 0) as shots_hit",
			"coalesce(sum(shots_fired), 0) as shots_fired",
			"floor(coalesce(sum(playtime_seconds), 0) / 3600) as total_playtime",
		}

		conds = append(conds, fmt.Sprintf("user_id IN (%v)", util.IntArrayToString(userData.Ids, ",")))

		conds = append(conds, "period BETWEEN yearweek(?) AND yearweek(?)")
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

		if len(req.ServerIds) > 0 {
			conds = append(conds, fmt.Sprintf("server_id IN (%v)", util.IntArrayToString(req.ServerIds, ",")))
		}

		if req.Perk != 0 {
			conds = append(conds, "perk = ?")
			args = append(args, req.Perk)

			fields = append(fields,
				"coalesce(sum(zedtime_length), 0) as zedtime_length",
				"coalesce(sum(zedtime_count), 0) as zedtime_count",
				"coalesce(sum(buffs_active_length), 0) as buffs_active_length",
				"coalesce(sum(buffs_total_length), 0) as buffs_total_length",
			)
		}

		sq = fmt.Sprintf(`
			SELECT user_id, %v
			FROM %v
			WHERE %v
			GROUP BY user_id
		`, strings.Join(fields, ", "), tableName, strings.Join(conds, " AND "))
	}

	stmt := fmt.Sprintf(`
		SELECT %v
		FROM (%v) t
		INNER JOIN users ON users.id = t.user_id
		GROUP BY user_id
		ORDER BY FIELD(users.id, %v)
		`, strings.Join(fields, ", "), sq, util.IntArrayToString(userData.Ids, ","),
	)

	rows, err := s.db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	res := LeaderBoardsResponse{
		Items: []*LeaderBoardsResponseItem{},
		Metadata: &models.PaginationResponse{
			Page:           req.Page,
			ResultsPerPage: 50,
			TotalResults:   userData.Total,
		},
	}
	steamIdSet := make(map[string]bool)

	for rows.Next() {
		item := LeaderBoardsResponseItem{}

		bindings := []any{
			&item.Id, &item.Name,
			&item.Type, &item.AuthId,
			&item.TotalGames, &item.TotalDeaths,
			&item.TotalDamage, &item.TotalKills,
			&item.TotalLargeKills, &item.TotalHeals,
			&item.TotalPlaytime,
		}

		if req.Perk != 0 {
			bindings = append(bindings,
				&item.HSAccuracy, &item.Accuracy,
				&item.AverageZedtime, &item.AverageBuffsUptime,
			)
		}

		err := rows.Scan(bindings...)
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
		args := make([]any, 0)

		conds = append(conds, "period BETWEEN yearweek(?) AND yearweek(?)")
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

		if req.Perk != 0 {
			conds = append(conds, "perk = ?")
			args = append(args, req.Perk)
		}

		if len(req.ServerIds) > 0 {
			conds = append(conds, fmt.Sprintf("server_id IN (%v)", util.IntArrayToString(req.ServerIds, ",")))
		}

		conds = append(conds, fmt.Sprintf("user_id IN (%v)", util.IntArrayToString(userData.Ids, ",")))

		stmt := fmt.Sprintf(`
			SELECT
				user_id,
				ANY_VALUE(max_damage_session_id) as session_id,
				max(max_damage) as metric
			FROM (
				SELECT
					user_id,
					max_damage_session_id,
					max(max_damage) as max_damage
				FROM %v
				WHERE %v
				GROUP BY user_id, max_damage_session_id
				ORDER BY max_damage DESC
			) t
			GROUP BY user_id`, tableName, strings.Join(conds, " AND "),
		)

		rows, err := s.db.Query(stmt, args...)
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

func (s *LeaderBoardsService) getMostDamageTop(
	req LeaderBoardsRequest,
) (*userIdResponse, error) {
	conds := make([]string, 0)
	args := make([]any, 0)

	conds = append(conds, "period BETWEEN yearweek(?) AND yearweek(?)")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	if len(req.ServerIds) > 0 {
		conds = append(conds, fmt.Sprintf("server_id IN (%v)", util.IntArrayToString(req.ServerIds, ",")))
	}

	tableName := "user_weekly_stats_total"
	if req.Perk != 0 {
		tableName = "user_weekly_stats_perk"
		conds = append(conds, "perk = ?")
		args = append(args, req.Perk)
	}

	stmt := fmt.Sprintf(`
		SELECT
			user_id,
			max(max_damage) as metric
		FROM (
			SELECT
				user_id,
				max_damage_session_id,
				max(max_damage) as max_damage
			FROM %v
			WHERE %v
			GROUP BY user_id, max_damage_session_id
			ORDER BY max_damage DESC
		) t
		GROUP BY user_id
		ORDER BY metric desc
		LIMIT %v, 50`, tableName, strings.Join(conds, " AND "), req.Page*50,
	)

	rows, err := s.db.Query(stmt, args...)
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

	totalRows, err := s.getTotalRows(req, false)
	if err != nil {
		return nil, err
	}

	return &userIdResponse{
		Ids:   items,
		Total: totalRows,
	}, nil
}

func (s *LeaderBoardsService) getLeaderboardIds(
	req LeaderBoardsRequest,
) (*userIdResponse, error) {
	conds := make([]string, 0)
	args := make([]any, 0)

	var metric string
	switch req.OrderBy {
	case TotalGames:
		metric = "sum(total_games) as metric"
	case TotalDeaths:
		metric = "coalesce(sum(deaths), 0) as metric"
	case TotalDamage:
		metric = "coalesce(sum(damage_dealt), 0) as metric"
	case TotalKills:
		metric = "coalesce(sum(total_kills), 0) as metric"
	case TotalLargeKills:
		metric = "coalesce(sum(large_kills), 0) as metric"
	case TotalHeals:
		metric = "coalesce(sum(heals_given), 0) as metric"
	case TotalPlaytime:
		metric = "floor(coalesce(sum(playtime_seconds), 0) / 3600) as metric"
	case AverageZedtime:
		metric = "coalesce(sum(zedtime_length) / sum(zedtime_count), 0) as metric"
	case AverageBuffsUptime:
		metric = "coalesce(sum(buffs_active_length) / sum(buffs_total_length), 0) as metric"
	case Accuracy:
		metric = "greatest(0, least(coalesce(sum(shots_hit) / sum(shots_fired), 0), 1)) as metric"
	case HsAccuracy:
		metric = "greatest(0, least(coalesce(sum(shots_hs) / sum(shots_hit), 0), 1)) as metric"
	}

	conds = append(conds, "period BETWEEN yearweek(?) AND yearweek(?)")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	if len(req.ServerIds) > 0 {
		conds = append(conds, fmt.Sprintf("server_id IN (%v)", util.IntArrayToString(req.ServerIds, ",")))
	}

	tableName := "user_weekly_stats_total"
	if req.Perk != 0 {
		tableName = "user_weekly_stats_perk"
		conds = append(conds, "perk = ?")
		args = append(args, req.Perk)
	}

	restrictByGamesCond := ""
	if req.To.Sub(req.From).Hours()/24 >= 81 {
		// 3 Month leaderboard requires at least 25 recent games
		restrictByGamesCond = "HAVING sum(total_games) >= 25"
	} else if req.To.Sub(req.From).Hours()/24 >= 28 {
		// Month leaderboard requires at least 10 recent games
		restrictByGamesCond = "HAVING sum(total_games) >= 10"
	} else {
		restrictByGamesCond = "HAVING sum(total_games) >= 3"
	}

	stmt := fmt.Sprintf(`
		SELECT users.id, t.metric
		FROM (
			SELECT user_id, %v
			FROM %v
			WHERE %v
			GROUP BY user_id
			%v
			ORDER BY metric DESC
			LIMIT %v, 50
		) t
		INNER JOIN users ON users.id = t.user_id
		ORDER BY metric DESC, name ASC
		`, metric, tableName, strings.Join(conds, " AND "), restrictByGamesCond, req.Page*50,
	)

	rows, err := s.db.Query(stmt, args...)
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

	totalRows, err := s.getTotalRows(req, true)
	if err != nil {
		return nil, err
	}

	return &userIdResponse{
		Ids:   items,
		Total: totalRows,
	}, nil
}

func (s *LeaderBoardsService) validateRequest(req LeaderBoardsRequest) error {
	if req.Perk == 0 {
		switch req.OrderBy {
		case AverageZedtime:
		case AverageBuffsUptime:
			return fmt.Errorf("selected type requires selected perk")
		}
	}

	return nil
}

func (s *LeaderBoardsService) getTotalRows(
	req LeaderBoardsRequest, restrictByGames bool,
) (int, error) {
	conds := make([]string, 0)
	args := make([]any, 0)

	conds = append(conds, "period BETWEEN yearweek(?) AND yearweek(?)")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	tableName := "user_weekly_stats_total"
	if req.Perk != 0 {
		tableName = "user_weekly_stats_perk"
		conds = append(conds, "perk = ?")
		args = append(args, req.Perk)
	}

	if len(req.ServerIds) > 0 {
		conds = append(conds, fmt.Sprintf("server_id IN (%v)", util.IntArrayToString(req.ServerIds, ",")))
	}

	restrictByGamesCond := ""
	if restrictByGames {
		if req.To.Sub(req.From).Hours()/24 >= 81 {
			// 3 Month leaderboard requires at least 25 recent games
			restrictByGamesCond = "HAVING sum(total_games) >= 25"
		} else if req.To.Sub(req.From).Hours()/24 >= 28 {
			// Month leaderboard requires at least 10 recent games
			restrictByGamesCond = "HAVING sum(total_games) >= 10"
		} else {
			restrictByGamesCond = "HAVING sum(total_games) >= 3"
		}
	}

	var total int
	stmt := fmt.Sprintf(`
		SELECT count(*)
		FROM (
			SELECT user_id, sum(total_games) as metric
			FROM %v
			WHERE %v
			GROUP BY user_id
			%v
			ORDER BY metric DESC
		) t
		`, tableName, strings.Join(conds, " AND "), restrictByGamesCond,
	)

	row := s.db.QueryRow(stmt, args...)

	err := row.Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}
