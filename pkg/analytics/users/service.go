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
	args := []interface{}{req.UserId}

	if req.From != nil && req.To != nil {
		conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
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
		conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
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
	args := []interface{}{}

	conds = append(conds,
		"user_id = ?",
		"DATE(session.started_at) BETWEEN ? AND ?",
	)

	args = append(args, req.UserId)

	if req.From != nil && req.To != nil {
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	} else {
		args = append(args, "2000-01-01", "3000-01-01")
	}

	if req.Perk != nil {
		conds = append(conds, "perk = ?")
		args = append(args, *req.Perk)
	}

	sql := fmt.Sprintf(`
		SELECT
			DATE(session.started_at) as period,
			count(distinct session.id) as playtime_count,
			round(coalesce(sum(playtime_seconds), 0) / 60) as playtime_minutes
		FROM session
		INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
		WHERE %v
		GROUP BY period
		ORDER BY period`, strings.Join(conds, " AND "),
	)

	rows, err := s.db.Query(sql, args...)
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
	args := []interface{}{}

	conds = append(conds,
		"user_id = ?",
		"DATE(session.started_at) BETWEEN ? AND ?",
	)

	args = append(args, req.UserId)

	if req.From != nil && req.To != nil {
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	} else {
		args = append(args, "2000-01-01", "3000-01-01")
	}

	if req.Perk != nil {
		conds = append(conds, "perk = ?")
		args = append(args, *req.Perk)
	}

	sql := fmt.Sprintf(`
		SELECT
			DATE(session.started_at) as period,
			sum(shots_hit) / greatest(sum(shots_fired), 1) as accuracy,
			sum(shots_hs) / greatest(sum(shots_hit), 1) as hs_accuracy
		FROM session
		INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
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
	limit := 5
	if req.Limit != nil && *req.Limit > 5 {
		limit = *req.Limit
	}

	sql := fmt.Sprintf(`
		SELECT
			users.id as user_id,
			users.name as name,
			users.auth_type,
			users.auth_id,
			count(*) as games_played,
			coalesce(sum(status = 2), 0) as wins
		FROM (
			SELECT
				aggr.user_id,
				t.session_id,
				t.status
			FROM (
				SELECT 
					session.id as session_id,
					session.status as status
				FROM session
				INNER JOIN session_aggregated aggr ON aggr.session_id = session.id
				WHERE user_id = %v
				GROUP BY session.id
			) t
			INNER JOIN session_aggregated aggr ON aggr.session_id = t.session_id
			WHERE user_id != %v
			GROUP BY aggr.user_id, t.session_id
		) t
		INNER JOIN users ON users.id = t.user_id
		GROUP BY user_id
		ORDER BY games_played desc, name asc
		LIMIT %v`, req.UserId, req.UserId, limit,
	)

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	res := GetTeammatesResponse{
		Items: []*GetTeammatesResponseItem{},
	}
	steamIdSet := make(map[string]bool)

	for rows.Next() {
		item := GetTeammatesResponseItem{}

		err := rows.Scan(
			&item.Id, &item.Name,
			&item.Type, &item.AuthId,
			&item.Games, &item.Wins,
		)
		if err != nil {
			return nil, err
		}

		if item.Type == models.Steam {
			steamIdSet[item.AuthId] = true
		}

		res.Items = append(res.Items, &item)
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

	return &res, nil
}

func (s *UserAnalyticsService) getPlayedMaps(
	req GetPlayedMapsRequest,
) (*GetPlayedMapsResponse, error) {
	conds := []string{}
	args := []any{}

	conds = append(conds,
		"player_id = ?",
		"DATE(session.started_at) BETWEEN ? AND ?",
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

	conds := []string{}
	args := []any{}

	conds = append(conds,
		"player_id = ?",
		"DATE(session.started_at) BETWEEN ? AND ?",
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
			WHERE wsp.player_id != %v
			WINDOW w AS (PARTITION BY wsp.player_id ORDER BY wsp.id DESC)
			ORDER BY wsp.id DESC
		), user_played_with AS (
			SELECT 
				cte.user_id as user_id,
				cte.session_id as session_id,
				cte.wsp_id as wsp_id
			FROM user_rating cte
			WHERE rating = 1
			LIMIT %v, %v
		)
		SELECT DISTINCT
			users.id AS user_id,
			users.name AS user_name,
			users.auth_id AS auth_id,
			users.auth_type AS auth_type,
			server.id AS server_id,
			server.name AS server_name,
			maps.id AS map_id,
			maps.name AS map_name,
			cte.session_id AS session_id,
			user_wsp.perk AS user_perk,
			last_seen.created_at AS last_seen
		FROM user_played_with cte
		INNER JOIN session ON session.id = cte.session_id
		INNER JOIN server ON server.id = session.server_id
		INNER JOIN maps ON maps.id = session.map_id
		INNER JOIN wave_stats ws ON ws.session_id = cte.session_id
		INNER JOIN wave_stats_player user_wsp ON user_wsp.stats_id = ws.id AND user_wsp.player_id = %v
		INNER JOIN wave_stats_player last_seen ON last_seen.id = cte.wsp_id
		INNER JOIN users ON users.id = cte.user_id
		ORDER BY last_seen DESC, user_id ASC
		`, strings.Join(conds, " AND "), req.UserId, page*limit, limit, req.UserId,
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

	for rows.Next() {
		var perk int
		item := GetLastSeenUsersResponseItem{
			Server: ServerData{},
			Map:    MapData{},
			Perks:  []int{},
		}

		err := rows.Scan(
			&item.Id, &item.Name,
			&item.AuthId, &item.Type,
			&item.Server.Id, &item.Server.Name,
			&item.Map.Id, &item.Map.Name,
			&item.SessionId, &perk,
			&item.LastSeen,
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

	{
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
				WHERE wsp.player_id != %v
				WINDOW w AS (PARTITION BY wsp.player_id ORDER BY wsp.id DESC)
				ORDER BY wsp.id DESC
			)
			SELECT count(*)
			FROM user_rating cte
			WHERE rating = 1
			`, strings.Join(conds, " AND "), req.UserId,
		)

		row := s.db.QueryRow(stmt, args...)

		var count int
		row.Scan(&count)
		res.Metadata.TotalResults = count
	}

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
		"DATE(session.started_at) BETWEEN ? AND ?",
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
			LIMIT %v, %v
		)
		SELECT DISTINCT
			cte.session_id AS session_id,
			user_wsp.perk AS user_perk,
			server.id AS server_id,
			server.name AS server_name,
			maps.id AS map_id,
			maps.name AS map_name,
			max(user_wsp.created_at) OVER w AS last_seen
		FROM user_played_with cte
		INNER JOIN session ON session.id = cte.session_id
		INNER JOIN server ON server.id = session.server_id
		INNER JOIN maps ON maps.id = session.map_id
		INNER JOIN wave_stats ws ON ws.session_id = cte.session_id
		INNER JOIN wave_stats_player user_wsp ON user_wsp.stats_id = ws.id AND user_wsp.player_id = %v
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

	for rows.Next() {
		var perk int
		item := GetLastSessionsWithUserResponseItem{
			Server: ServerData{},
			Map:    MapData{},
			Perks:  []int{},
		}

		err := rows.Scan(
			&item.SessionId, &perk,
			&item.Server.Id, &item.Server.Name,
			&item.Map.Id, &item.Map.Name,
			&item.LastSeen,
		)
		if err != nil {
			return nil, err
		}

		if len(res.Items) > 0 && res.Items[len(res.Items)-1].SessionId == item.SessionId {
			last := res.Items[len(res.Items)-1]
			last.Perks = append(last.Perks, perk)
		} else {
			item.Perks = append(item.Perks, perk)
			res.Items = append(res.Items, &item)
		}
	}

	{
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
			)
			SELECT count(*) 
			FROM user_sessions t1
			INNER JOIN other_user_sessions t2 ON t1.session_id = t2.session_id
			`, strings.Join(conds, " AND "), req.OtherUserId,
		)

		row := s.db.QueryRow(stmt, args...)

		var count int
		row.Scan(&count)
		res.Metadata.TotalResults = count
	}

	return &res, nil
}
