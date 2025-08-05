package users

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
	"github.com/theggv/kf2-stats-backend/pkg/users"
)

type UserAnalyticsService struct {
	db *sql.DB

	userService *users.UserService
}

func NewUserAnalyticsService(db *sql.DB) *UserAnalyticsService {
	service := UserAnalyticsService{
		db: db,
	}

	return &service
}

func (s *UserAnalyticsService) Inject(
	userService *users.UserService,
) {
	s.userService = userService
}

func (s *UserAnalyticsService) GetUserAnalytics(
	req UserAnalyticsRequest,
) (*UserAnalyticsResponse, error) {
	res := UserAnalyticsResponse{}

	conds := []string{"aggr.user_id = ?"}
	args := []any{req.UserId}

	if req.From != nil && req.To != nil {
		conds = append(conds, "DATE(session.updated_at) BETWEEN ? AND ?")
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	}

	sql := fmt.Sprintf(`
		SELECT
			count(distinct session.id) as total_games,
			coalesce(sum(kills.total), 0) as total_kills,
			coalesce(sum(deaths), 0) as total_deaths,
			floor(coalesce(sum(playtime_seconds), 0) / 60) as total_minutes
		FROM session
		INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
		INNER JOIN session_aggregated_kills kills ON aggr.id = kills.id
		WHERE %v`, strings.Join(conds, " AND "),
	)

	err := s.db.QueryRow(sql, args...).Scan(&res.Games, &res.Kills, &res.Deaths, &res.Minutes)
	if err != nil {
		return nil, err
	}

	sql = fmt.Sprintf(`
		SELECT count(distinct session.id) as total_wins
		FROM session
		INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
		WHERE %v AND status = 2
		`, strings.Join(conds, " AND "),
	)

	err = s.db.QueryRow(sql, args...).Scan(&res.Wins)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *UserAnalyticsService) GetPerksAnalytics(
	req UserPerksAnalyticsRequest,
) (*UserPerksAnalyticsResponse, error) {
	conds := []string{}
	args := []interface{}{}

	if req.From != nil && req.To != nil {
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	} else {
		args = append(args, "2000-01-01", "3000-01-01")
		args = append(args, "2000-01-01", "3000-01-01")
		args = append(args, "2000-01-01", "3000-01-01")
	}

	conds = append(conds, "aggr.user_id = ?", "aggr.perk > 0")
	args = append(args, req.UserId)

	if req.From != nil && req.To != nil {
		conds = append(conds, "DATE(session.updated_at) BETWEEN ? AND ?")
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	}

	sql := fmt.Sprintf(`
		SELECT 
			perk,
			total_games,
			get_user_games_count_by_perk(user_id, perk, 2, ?, ?) as total_wins,
			total_kills,
			large_kills,
			total_waves,
			total_deaths,
			get_avg_acc(user_id, perk, ?, ?) as accuracy,
			get_avg_hs_acc(user_id, perk, ?, ?) as hs_accuracy,
			heals_given,
			damage_dealt,
			damage_taken,
			total_minutes
		FROM (
			SELECT
				perk,
				min(user_id) as user_id,
				count(*) as total_games,
				sum(kills.total) as total_kills,
				sum(kills.large) as large_kills,
				sum(waves_played) as total_waves,
				sum(deaths) as total_deaths,
				sum(heals_given) as heals_given,
				sum(damage_dealt) as damage_dealt,
				sum(damage_taken) as damage_taken,
				floor(coalesce(sum(playtime_seconds), 0) / 60) as total_minutes
			FROM session
			INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
			INNER JOIN session_aggregated_kills kills ON aggr.id = kills.id
			WHERE %v
			GROUP BY perk
		) t
		WHERE total_kills > 0
		ORDER BY perk`, strings.Join(conds, " AND "),
	)

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	res := UserPerksAnalyticsResponse{
		Items: []UserPerksAnalyticsResponseItem{},
	}

	for rows.Next() {
		item := UserPerksAnalyticsResponseItem{}

		err := rows.Scan(
			&item.Perk, &item.Games, &item.Wins,
			&item.Kills, &item.LargeKills,
			&item.Waves, &item.Deaths,
			&item.Accuracy, &item.HSAccuracy,
			&item.HealsGiven, &item.DamageDealt, &item.DamageTaken,
			&item.Minutes,
		)

		if err != nil {
			return nil, err
		}

		res.Items = append(res.Items, item)
	}

	averageZt, err := s.getAverageZedtime(req)
	if err != nil {
		return nil, err
	}

	res.AverageZedtime = averageZt

	return &res, nil
}

func (s *UserAnalyticsService) getAverageZedtime(
	req UserPerksAnalyticsRequest) (float64, error) {

	args := []any{}
	args = append(args, req.UserId)

	if req.From != nil && req.To != nil {
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	} else {
		args = append(args, "2000-01-01", "3000-01-01")
	}

	stmt := `SELECT get_avg_zt(?, ?, ?)`

	var averageZt float64
	err := s.db.QueryRow(stmt, args...).Scan(&averageZt)

	return averageZt, err
}

func (s *UserAnalyticsService) getPlaytimeHist(
	req UserPerkHistRequest,
) (*PlayTimeHist, error) {
	conds := []string{}
	args := []any{}

	conds = append(conds, "aggr.user_id = ?")
	args = append(args, req.UserId)

	if req.AuthUser != nil && req.AuthUser.UserId == req.UserId {
		conds = append(conds, "DATE(session.updated_at) BETWEEN ? AND ?")
		if req.From != nil && req.To != nil {
			args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
		} else {
			args = append(args, "2000-01-01", "3000-01-01")
		}
	} else {
		conds = append(conds, `
			DATE(session.updated_at) BETWEEN 
				CURRENT_TIMESTAMP - INTERVAL 180 DAY AND CURRENT_TIMESTAMP`,
		)
	}

	if len(req.Perks) > 0 {
		conds = append(conds, fmt.Sprintf("aggr.perk IN (%v)", util.IntArrayToString(req.Perks, ",")))
	}

	if len(req.MapIds) > 0 {
		conds = append(conds, fmt.Sprintf("map_id IN (%v)", util.IntArrayToString(req.MapIds, ",")))
	}

	if len(req.ServerIds) > 0 {
		conds = append(conds, fmt.Sprintf("server_id IN (%v)", util.IntArrayToString(req.ServerIds, ",")))
	}

	if req.Mode != nil {
		conds = append(conds, fmt.Sprintf("session.mode = %v", *req.Mode))
	}

	if req.Mode == nil || *req.Mode != models.ControlledDifficulty {
		if req.Difficulty != nil {
			conds = append(conds, fmt.Sprintf("session.diff = %v", *req.Difficulty))
		}
	}

	if req.Status != nil {
		conds = append(conds, fmt.Sprintf("session.status = %v", *req.Status))
	}

	if req.Length != nil {
		conds = append(conds, fmt.Sprintf("session.length = %v", *req.Length))
	}

	if req.MinWave != nil {
		conds = append(conds, fmt.Sprintf("gd.wave >= %v", *req.MinWave))
	}

	if req.Mode != nil && *req.Mode == models.ControlledDifficulty {
		if req.SpawnCycle != nil {
			conds = append(conds, "cd.spawn_cycle LIKE ?")
			args = append(args, fmt.Sprintf("%%%v%%", *req.SpawnCycle))
		}

		if req.ZedsType != nil {
			conds = append(conds, "cd.zeds_type LIKE ?")
			args = append(args, fmt.Sprintf("%v%%", *req.ZedsType))
		}

		if req.MaxMonsters != nil {
			conds = append(conds, fmt.Sprintf("cd.max_monsters = %v", *req.MaxMonsters))
		}
	}

	stmt := fmt.Sprintf(`
		SELECT
			DATE(session.updated_at) as period,
			count(distinct session.id) as playtime_count,
			round(coalesce(sum(playtime_seconds), 0) / 60) as playtime_minutes
		FROM session
		INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
		INNER JOIN session_game_data gd on gd.session_id = session.id
		LEFT JOIN session_game_data_extra cd on cd.session_id = session.id
		WHERE %v
		GROUP BY period
		ORDER BY period`, strings.Join(conds, " AND "),
	)

	rows, err := s.db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	res := PlayTimeHist{
		Items: []PlayTimeHistItem{},
	}

	for rows.Next() {
		item := PlayTimeHistItem{}

		err := rows.Scan(&item.Period, &item.Count, &item.Minutes)
		if err != nil {
			return nil, err
		}

		res.Items = append(res.Items, item)
	}

	return &res, nil
}

func (s *UserAnalyticsService) getAccuracyHist(
	req UserPerkHistRequest,
) (*AccuracyHist, error) {
	conds := []string{}
	args := []any{}

	conds = append(conds, "aggr.user_id = ?")
	args = append(args, req.UserId)

	if len(req.Perks) > 0 {
		conds = append(conds, fmt.Sprintf("aggr.perk IN (%v)", util.IntArrayToString(req.Perks, ",")))
	}

	if len(req.MapIds) > 0 {
		conds = append(conds, fmt.Sprintf("map_id IN (%v)", util.IntArrayToString(req.MapIds, ",")))
	}

	if len(req.ServerIds) > 0 {
		conds = append(conds, fmt.Sprintf("server_id IN (%v)", util.IntArrayToString(req.ServerIds, ",")))
	}

	if req.Mode != nil {
		conds = append(conds, fmt.Sprintf("session.mode = %v", *req.Mode))
	}

	if req.Mode == nil || *req.Mode != models.ControlledDifficulty {
		if req.Difficulty != nil {
			conds = append(conds, fmt.Sprintf("session.diff = %v", *req.Difficulty))
		}
	}

	if req.Status != nil {
		conds = append(conds, fmt.Sprintf("session.status = %v", *req.Status))
	}

	if req.Length != nil {
		conds = append(conds, fmt.Sprintf("session.length = %v", *req.Length))
	}

	if req.MinWave != nil {
		conds = append(conds, fmt.Sprintf("gd.wave >= %v", *req.MinWave))
	}

	if req.Mode != nil && *req.Mode == models.ControlledDifficulty {
		if req.SpawnCycle != nil {
			conds = append(conds, "cd.spawn_cycle LIKE ?")
			args = append(args, fmt.Sprintf("%%%v%%", *req.SpawnCycle))
		}

		if req.ZedsType != nil {
			conds = append(conds, "cd.zeds_type LIKE ?")
			args = append(args, fmt.Sprintf("%v%%", *req.ZedsType))
		}

		if req.MaxMonsters != nil {
			conds = append(conds, fmt.Sprintf("cd.max_monsters = %v", *req.MaxMonsters))
		}
	}

	sql := fmt.Sprintf(`
		SELECT
			DATE(session.updated_at) as period,
			sum(shots_hit) / greatest(sum(shots_fired), 1) as accuracy,
			sum(shots_hs) / greatest(sum(shots_hit), 1) as hs_accuracy
		FROM session
		INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
		INNER JOIN session_game_data gd on gd.session_id = session.id
		LEFT JOIN session_game_data_extra cd on cd.session_id = session.id
		WHERE %v
		GROUP BY period
		ORDER BY period`, strings.Join(conds, " AND "),
	)

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	res := AccuracyHist{
		Items: []AccuracyHistItem{},
	}

	for rows.Next() {
		item := AccuracyHistItem{}

		err := rows.Scan(&item.Period, &item.Accuracy, &item.HSAccuracy)
		if err != nil {
			return nil, err
		}

		res.Items = append(res.Items, item)
	}

	return &res, nil
}

func (s *UserAnalyticsService) getTeammates(
	req GetTeammatesRequest,
) (*GetTeammatesResponse, error) {
	page, limit := util.ParsePagination(req.Pager)

	if req.AuthUser == nil || req.AuthUser.UserId != req.UserId {
		page = 0
		limit = 5
	}

	stmt := fmt.Sprintf(`
		WITH user_sessions AS (
			SELECT DISTINCT session_id
			FROM session_aggregated aggr
			WHERE user_id = %v
		), user_played_with AS (
			SELECT DISTINCT 
				aggr.user_id as user_id,
				cte.session_id as session_id
			FROM user_sessions cte
			INNER JOIN session_aggregated aggr ON aggr.session_id = cte.session_id
			WHERE user_id != %v
		), user_stats AS (
			SELECT DISTINCT
				cte.user_id as user_id,
				count(session.id) OVER w as total_games,
				count(CASE WHEN session.status = 2 THEN 1 END) OVER w as total_wins
			FROM user_played_with cte
			INNER JOIN session ON session.id = cte.session_id
			WINDOW w AS (partition by cte.user_id)
			ORDER BY total_games DESC, user_id ASC
		), pagination AS (
			SELECT *
			FROM user_stats cte
			LIMIT %v, %v
		), metadata AS (
			SELECT count(*) as count
			FROM user_stats cte
		)
		SELECT 
			users.id as user_id,
			users.name as user_name,
			users.auth_type as auth_type,
			users.auth_id as auth_id,
			cte.total_games,
			cte.total_wins,
			metadata.count as total_results
		FROM pagination cte
		INNER JOIN users ON users.id = cte.user_id
		CROSS JOIN metadata
		`, req.UserId, req.UserId, page*limit, limit,
	)

	rows, err := s.db.Query(stmt)
	if err != nil {
		return nil, err
	}

	res := GetTeammatesResponse{
		Items: []*GetTeammatesResponseItem{},
		Metadata: &models.PaginationResponse{
			Page:           page,
			ResultsPerPage: limit,
		},
	}
	steamIdSet := make(map[string]bool)

	var count int
	for rows.Next() {
		item := GetTeammatesResponseItem{}

		err := rows.Scan(
			&item.Id, &item.Name,
			&item.Type, &item.AuthId,
			&item.Games, &item.Wins,
			&count,
		)
		if err != nil {
			return nil, err
		}

		if item.Type == models.Steam {
			steamIdSet[item.AuthId] = true
		}

		res.Items = append(res.Items, &item)
	}

	res.Metadata.TotalResults = count

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

func (s *UserAnalyticsService) getPlayedMaps(
	req GetPlayedMapsRequest,
) (*GetPlayedMapsResponse, error) {
	conds := []string{}
	args := []any{}

	conds = append(conds,
		"player_id = ?",
		"DATE(session.updated_at) BETWEEN ? AND ?",
		"session.completed_at IS NOT NULL",
	)

	args = append(args, req.UserId)

	if req.From != nil && req.To != nil {
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	} else {
		args = append(args, "2000-01-01", "3000-01-01")
	}

	if len(req.Perks) > 0 {
		conds = append(conds, fmt.Sprintf("perk IN (%v)",
			util.IntArrayToString(req.Perks, ",")),
		)
	}

	if len(req.ServerIds) > 0 {
		conds = append(conds, fmt.Sprintf("server_id IN (%v)",
			util.IntArrayToString(req.ServerIds, ",")),
		)
	}

	stmt := fmt.Sprintf(`
		WITH cte AS (
			SELECT DISTINCT session.id AS session_id
			FROM session
			INNER JOIN wave_stats ws ON ws.session_id = session.id
			INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
			WHERE %v
		)
		SELECT DISTINCT
			maps.name AS map_name,
			COUNT(session.id) over w AS total_games,
			COUNT(CASE WHEN session.status = 2 THEN 1 END) over w AS total_wins,
			MAX(session.completed_at) over w AS last_played
		FROM cte
		INNER JOIN session ON cte.session_id = session.id
		INNER JOIN maps ON maps.id = session.map_id
		WINDOW w AS (PARTITION BY maps.id)
		ORDER BY total_games DESC
		`, strings.Join(conds, " AND "),
	)

	rows, err := s.db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	res := GetPlayedMapsResponse{
		Items: []*GetPlayedMapsResponseItem{},
	}
	for rows.Next() {
		item := GetPlayedMapsResponseItem{}

		err := rows.Scan(
			&item.Name, &item.TotalGames, &item.TotalWins, &item.LastPlayed,
		)
		if err != nil {
			return nil, err
		}

		res.Items = append(res.Items, &item)
	}

	return &res, nil
}

func (s *UserAnalyticsService) getLastSeenUsers(
	req GetLastSeenUsersRequest,
) (*GetLastSeenUsersResponse, error) {
	page, limit := util.ParsePagination(req.Pager)

	userSessionConds := []string{}
	args := []any{}

	userSessionConds = append(userSessionConds,
		"player_id = ?",
		"DATE(session.updated_at) BETWEEN ? AND ?",
	)

	args = append(args, req.UserId)

	if req.From != nil && req.To != nil {
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	} else {
		args = append(args, "2000-01-01", "3000-01-01")
	}

	if len(req.Perks) > 0 {
		userSessionConds = append(userSessionConds, fmt.Sprintf("perk IN (%v)",
			util.IntArrayToString(req.Perks, ",")),
		)
	}

	if len(req.ServerIds) > 0 {
		userSessionConds = append(userSessionConds, fmt.Sprintf("server_id IN (%v)",
			util.IntArrayToString(req.ServerIds, ",")),
		)
	}

	userRatingConds := []string{"wsp.player_id != ?"}
	args = append(args, req.UserId)

	req.SearchText = strings.TrimSpace(req.SearchText)
	if len(req.SearchText) > 0 {
		userRatingConds = append(userRatingConds, "(LOWER(users.name) LIKE ? OR users.auth_id = ?)")
		args = append(args, fmt.Sprintf("%%%v%%", req.SearchText), req.SearchText)
	}

	stmt := fmt.Sprintf(`
		WITH user_sessions AS (
			SELECT DISTINCT session.id as session_id
			FROM session
			INNER JOIN wave_stats ws ON ws.session_id = session.id
			INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
			WHERE %v
		), user_rating AS (
			SELECT 
				wsp.player_id AS user_id,
				ws.session_id AS session_id,
				wsp.stats_id AS stats_id,
				wsp.id AS wsp_id,        
				DENSE_RANK() OVER w AS rating
			FROM user_sessions cte
			INNER JOIN wave_stats ws ON ws.session_id = cte.session_id
			INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
			INNER JOIN users ON users.id = wsp.player_id
			WHERE %v
			WINDOW w AS (PARTITION BY wsp.player_id ORDER BY wsp.id DESC)
			ORDER BY wsp.id DESC
		), user_played_with AS (
			SELECT 
				cte.user_id as user_id,
				cte.session_id as session_id,
				cte.wsp_id as wsp_id
			FROM user_rating cte
			WHERE rating = 1
		), pagination AS (
			SELECT *
			FROM user_played_with cte
			LIMIT %v, %v
		), metadata AS (
			SELECT count(*) as count
			FROM user_played_with cte
		)
		SELECT DISTINCT
			session.id AS session_id,
			session.status AS session_status,
			session.diff AS session_diff,
			session.mode AS sesion_mode,
			session.length AS session_length,
			user_wsp.perk AS user_perk,
			users.id AS user_id,
			users.name AS user_name,
			users.auth_id AS auth_id,
			users.auth_type AS auth_type,
			server.id AS server_id,
			server.name AS server_name,
			maps.id AS map_id,
			maps.name AS map_name,
			last_seen.created_at AS last_seen,
			metadata.count AS count
		FROM pagination cte
		INNER JOIN session ON session.id = cte.session_id
		INNER JOIN server ON server.id = session.server_id
		INNER JOIN maps ON maps.id = session.map_id
		INNER JOIN wave_stats ws ON ws.session_id = cte.session_id
		INNER JOIN wave_stats_player user_wsp ON user_wsp.stats_id = ws.id AND user_wsp.player_id = %v
		INNER JOIN wave_stats_player last_seen ON last_seen.id = cte.wsp_id
		INNER JOIN users ON users.id = cte.user_id
		CROSS JOIN metadata
		ORDER BY last_seen DESC, user_id ASC
		`,
		strings.Join(userSessionConds, " AND "),
		strings.Join(userRatingConds, " AND "),
		page*limit, limit, req.UserId,
	)

	rows, err := s.db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	res := GetLastSeenUsersResponse{
		Items: []*GetLastSeenUsersResponseItem{},
		Metadata: &models.PaginationResponse{
			Page:           page,
			ResultsPerPage: limit,
		},
	}
	steamIdSet := make(map[string]bool)

	var count int
	for rows.Next() {
		var perk int
		item := GetLastSeenUsersResponseItem{
			Session: SessionData{},
			Server:  ServerData{},
			Map:     MapData{},
			Perks:   []int{},
		}

		err := rows.Scan(
			&item.Session.Id, &item.Session.Status,
			&item.Session.Difficulty, &item.Session.Mode, &item.Session.Length,
			&perk,
			&item.Id, &item.Name,
			&item.AuthId, &item.Type,
			&item.Server.Id, &item.Server.Name,
			&item.Map.Id, &item.Map.Name,
			&item.LastSeen, &count,
		)
		if err != nil {
			return nil, err
		}

		if item.Type == models.Steam {
			steamIdSet[item.AuthId] = true
		}

		if len(res.Items) > 0 && res.Items[len(res.Items)-1].Id == item.Id {
			last := res.Items[len(res.Items)-1]
			last.Perks = append(last.Perks, perk)
		} else {
			item.Perks = append(item.Perks, perk)
			res.Items = append(res.Items, &item)
		}
	}

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

	res.Metadata.TotalResults = count

	return &res, nil
}

func (s *UserAnalyticsService) getLastGamesWithUser(
	req GetLastSessionsWithUserRequest,
) (*GetLastSessionsWithUserResponse, error) {
	page, limit := util.ParsePagination(req.Pager)

	conds := []string{}
	args := []any{}

	conds = append(conds,
		"player_id = ?",
		"DATE(session.updated_at) BETWEEN ? AND ?",
	)

	args = append(args, req.UserId)

	if req.From != nil && req.To != nil {
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	} else {
		args = append(args, "2000-01-01", "3000-01-01")
	}

	if len(req.Perks) > 0 {
		conds = append(conds, fmt.Sprintf("perk IN (%v)",
			util.IntArrayToString(req.Perks, ",")),
		)
	}

	if len(req.ServerIds) > 0 {
		conds = append(conds, fmt.Sprintf("server_id IN (%v)",
			util.IntArrayToString(req.ServerIds, ",")),
		)
	}

	stmt := fmt.Sprintf(`
		WITH user_sessions AS (
			SELECT DISTINCT session.id as session_id
			FROM session
			INNER JOIN wave_stats ws ON ws.session_id = session.id
			INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
			WHERE %v
		), other_user_sessions AS (
			SELECT DISTINCT session.id as session_id
			FROM session
			INNER JOIN wave_stats ws ON ws.session_id = session.id
			INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
			WHERE wsp.player_id = %v
		), user_played_with AS (
			SELECT t1.session_id as session_id 
			FROM user_sessions t1
			INNER JOIN other_user_sessions t2 ON t1.session_id = t2.session_id
			ORDER BY session_id DESC
		), pagination AS (
			SELECT *
			FROM user_played_with cte
			LIMIT %v, %v
		), metadata AS (
			SELECT count(*) as count
			FROM user_played_with cte
		)
		SELECT DISTINCT
			session.id AS session_id,
			session.status AS session_status,
			session.diff AS session_diff,
			session.mode AS sesion_mode,
			session.length AS session_length,
			user_wsp.perk AS user_perk,
			server.id AS server_id,
			server.name AS server_name,
			maps.id AS map_id,
			maps.name AS map_name,
			max(user_wsp.created_at) OVER w AS last_seen,
			metadata.count AS count
		FROM pagination cte
		INNER JOIN session ON session.id = cte.session_id
		INNER JOIN server ON server.id = session.server_id
		INNER JOIN maps ON maps.id = session.map_id
		INNER JOIN wave_stats ws ON ws.session_id = cte.session_id
		INNER JOIN wave_stats_player user_wsp ON user_wsp.stats_id = ws.id AND user_wsp.player_id = %v
		CROSS JOIN metadata
		WINDOW w AS (partition by session.id)
		ORDER BY last_seen DESC, user_perk ASC
		`, strings.Join(conds, " AND "), req.OtherUserId, page*limit, limit, req.UserId,
	)

	rows, err := s.db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	res := GetLastSessionsWithUserResponse{
		Items: []*GetLastSessionsWithUserResponseItem{},
		Metadata: &models.PaginationResponse{
			Page:           page,
			ResultsPerPage: limit,
		},
	}

	var count int
	for rows.Next() {
		var perk int
		item := GetLastSessionsWithUserResponseItem{
			Session: SessionData{},
			Server:  ServerData{},
			Map:     MapData{},
			Perks:   []int{},
		}

		err := rows.Scan(
			&item.Session.Id, &item.Session.Status,
			&item.Session.Difficulty, &item.Session.Mode, &item.Session.Length,
			&perk,
			&item.Server.Id, &item.Server.Name,
			&item.Map.Id, &item.Map.Name,
			&item.LastSeen, &count,
		)
		if err != nil {
			return nil, err
		}

		if len(res.Items) > 0 && res.Items[len(res.Items)-1].Session.Id == item.Session.Id {
			last := res.Items[len(res.Items)-1]
			last.Perks = append(last.Perks, perk)
		} else {
			item.Perks = append(item.Perks, perk)
			res.Items = append(res.Items, &item)
		}
	}

	res.Metadata.TotalResults = count

	return &res, nil
}

func (s *UserAnalyticsService) getUserSessions(
	req FindUserSessionsRequest,
) (*FindUserSessionsResponse, error) {
	page, limit := util.ParsePagination(req.Pager)

	sortBy := "updated_at"
	if req.SortBy.Field == "damage_dealt" {
		sortBy = "damage_dealt"
	}

	direction := "ASC"
	if req.SortBy.Direction == models.Desc {
		direction = "DESC"
	}

	conds := []string{}
	args := []any{}

	conds = append(conds, "aggr.user_id = ?")
	args = append(args, req.UserId)

	if req.AuthUser != nil && req.AuthUser.UserId == req.UserId {
		conds = append(conds, "DATE(session.updated_at) BETWEEN ? AND ?")
		if req.From != nil && req.To != nil {
			args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
		} else {
			args = append(args, "2000-01-01", "3000-01-01")
		}
	}

	if len(req.Perks) > 0 {
		conds = append(conds, fmt.Sprintf("aggr.perk IN (%v)", util.IntArrayToString(req.Perks, ",")))
	}

	if len(req.MapIds) > 0 {
		conds = append(conds, fmt.Sprintf("map_id IN (%v)", util.IntArrayToString(req.MapIds, ",")))
	}

	if len(req.ServerIds) > 0 {
		conds = append(conds, fmt.Sprintf("server_id IN (%v)", util.IntArrayToString(req.ServerIds, ",")))
	}

	if req.Mode != nil {
		conds = append(conds, fmt.Sprintf("session.mode = %v", *req.Mode))
	}

	if req.Mode == nil || *req.Mode != models.ControlledDifficulty {
		if req.Difficulty != nil {
			conds = append(conds, fmt.Sprintf("session.diff = %v", *req.Difficulty))
		}
	}

	if req.Status != nil {
		conds = append(conds, fmt.Sprintf("session.status = %v", *req.Status))
	}

	if req.Length != nil {
		conds = append(conds, fmt.Sprintf("session.length = %v", *req.Length))
	}

	if req.MinWave != nil {
		conds = append(conds, fmt.Sprintf("gd.wave >= %v", *req.MinWave))
	}

	if req.Mode != nil && *req.Mode == models.ControlledDifficulty {
		if req.SpawnCycle != nil {
			conds = append(conds, "cd.spawn_cycle LIKE ?")
			args = append(args, fmt.Sprintf("%%%v%%", *req.SpawnCycle))
		}

		if req.ZedsType != nil {
			conds = append(conds, "cd.zeds_type LIKE ?")
			args = append(args, fmt.Sprintf("%v%%", *req.ZedsType))
		}

		if req.MaxMonsters != nil {
			conds = append(conds, fmt.Sprintf("cd.max_monsters = %v", *req.MaxMonsters))
		}
	}

	sql := fmt.Sprintf(`
		WITH user_session AS (
			SELECT aggr.id AS aggr_id
			FROM session
			INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
			INNER JOIN session_game_data gd on gd.session_id = session.id
			LEFT JOIN session_game_data_extra cd on cd.session_id = session.id
			WHERE %v
		), pagination AS (
			SELECT DISTINCT
				session.id AS session_id,
				sum(aggr.damage_dealt) OVER w AS damage_dealt,
				session.updated_at AS updated_at
			FROM user_session cte
			INNER JOIN session_aggregated aggr ON aggr.id = cte.aggr_id
			INNER JOIN session ON session.id = aggr.session_id
			WINDOW w AS (partition by session.id)
			ORDER BY %v %v
			LIMIT %v, %v
		), metadata AS (
			SELECT count(distinct aggr.session_id) AS total_results
			FROM user_session cte
			INNER JOIN session_aggregated aggr ON aggr.id = cte.aggr_id
		), user_perks AS (
			SELECT DISTINCT
				aggr.session_id AS session_id,
				aggr.perk AS perk
			FROM pagination cte
			INNER JOIN session_aggregated aggr ON aggr.session_id = cte.session_id
			WHERE user_id = %v
		)
		SELECT DISTINCT
			session.id,
			session.mode,
			session.length,
			session.diff,
			session.status,

			extra.spawn_cycle,
			extra.max_monsters,
			extra.wave_size_fakes,
			extra.zeds_type,

			gd.wave,
			gd.zeds_left,
			user_perks.perk,
			cte.damage_dealt,
			
			server.id AS server_id,
			server.name AS server_name,

			maps.id AS map_id,
			maps.name AS map_name,

			session.updated_at,
			metadata.total_results
		FROM pagination cte
		INNER JOIN session ON session.id = cte.session_id
		INNER JOIN user_perks ON user_perks.session_id = session.id
		INNER JOIN server ON server.id = session.server_id
		INNER JOIN maps ON maps.id = session.map_id
		INNER JOIN session_game_data gd on gd.session_id = session.id
		LEFT JOIN session_game_data_extra extra on extra.session_id = session.id
		CROSS JOIN metadata
		`, strings.Join(conds, " AND "),
		sortBy, direction,
		page*limit, limit,
		req.UserId,
	)

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	var count int
	res := FindUserSessionsResponse{
		Items: []*FindUserSessionsResponseItem{},
		Metadata: &models.PaginationResponse{
			Page:           page,
			ResultsPerPage: limit,
		},
	}

	for rows.Next() {
		var perk int
		item := FindUserSessionsResponseItem{
			Session:  SessionData{},
			Server:   ServerData{},
			Map:      MapData{},
			GameData: FindUserSessionsResponseItemGameData{},
			Stats:    FindUserSessionsResponseItemStats{},
			Perks:    []int{},
		}
		extraGameData := models.ExtraGameData{}

		rows.Scan(
			&item.Session.Id, &item.Session.Mode, &item.Session.Length,
			&item.Session.Difficulty, &item.Session.Status,
			&extraGameData.SpawnCycle, &extraGameData.MaxMonsters,
			&extraGameData.WaveSizeFakes, &extraGameData.ZedsType,
			&item.GameData.Wave, &item.GameData.ZedsLeft, &perk,
			&item.Stats.DamageDealt,
			&item.Server.Id, &item.Server.Name,
			&item.Map.Id, &item.Map.Name,
			&item.UpdatedAt, &count,
		)

		if extraGameData.SpawnCycle != nil {
			item.ExtraGameData = &extraGameData
		}

		if len(res.Items) > 0 && res.Items[len(res.Items)-1].Session.Id == item.Session.Id {
			last := res.Items[len(res.Items)-1]
			last.Perks = append(last.Perks, perk)
		} else {
			item.Perks = append(item.Perks, perk)
			res.Items = append(res.Items, &item)
		}
	}

	res.Metadata.TotalResults = count

	return &res, nil
}
