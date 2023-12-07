package users

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
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

	args := []interface{}{}
	args = append(args, req.UserId)

	if req.From != nil && req.To != nil {
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	} else {
		args = append(args, "2000-01-01", "3000-01-01")
	}

	sql := fmt.Sprintf(`
		SELECT get_avg_zt(?, ?, ?)`,
	)

	var averageZt float64
	err := s.db.QueryRow(sql, args...).Scan(&averageZt)

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

func (s *UserAnalyticsService) GetUsersTop(
	req GetUsersTopRequest,
) (*GetUsersTopResponse, error) {
	switch req.Type {
	case AverageZedtime:
		return s.getZedtimeTop(req)
	case MostHsAccuracy:
		return s.getHSAccuracyTop(req)
	}

	conds := make([]string, 0)
	args := make([]interface{}, 0)

	var metric string
	switch req.Type {
	case MostKills:
		metric = "coalesce(sum(kills.total), 0) as metric"
	case MostDeaths:
		metric = "coalesce(sum(deaths), 0) as metric"
	case MostPlaytime:
		metric = "floor(coalesce(sum(playtime_seconds), 0) / 60) as metric"
	case MostDamageDealt:
		metric = "coalesce(sum(damage_dealt), 0) as metric"
	case MostHealsGiven:
		metric = "coalesce(sum(heals_given), 0) as metric"
	}

	if req.From != nil && req.To != nil {
		conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	}

	sql := fmt.Sprintf(`
		SELECT
			users.id,
			users.name,
			users.auth_type,
			users.auth_id,
			t.total_games,
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

	res := GetUsersTopResponse{
		Items: []*GetUsersTopResponseItem{},
	}
	steamIdSet := make(map[string]bool)

	for rows.Next() {
		item := GetUsersTopResponseItem{}

		err := rows.Scan(
			&item.Id, &item.Name,
			&item.Type, &item.AuthId,
			&item.Games, &item.Metric,
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

func (s *UserAnalyticsService) getHSAccuracyTop(
	req GetUsersTopRequest,
) (*GetUsersTopResponse, error) {
	conds := make([]string, 0)
	args := make([]interface{}, 0)

	if req.Perk == 0 {
		return nil, errors.New(fmt.Sprintf("perk cannot be null with selected type"))
	}

	if req.From == nil || req.To == nil {
		return nil, errors.New(fmt.Sprintf("incorrect date"))
	}

	args = append(args, req.Perk, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	conds = append(conds, "perk = ?")
	args = append(args, req.Perk)

	conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
	args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	sql := fmt.Sprintf(`
		SELECT
			users.id as user_id,
			users.name as name,
			users.auth_type,
			users.auth_id,
			total_games,
			get_avg_hs_acc(users.id, ?, ?, ?) as metric
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
		INNER JOIN users ON users.id = t.user_id
		ORDER BY metric DESC
		LIMIT 50`, strings.Join(conds, " AND "),
	)

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	res := GetUsersTopResponse{
		Items: []*GetUsersTopResponseItem{},
	}
	steamIdSet := make(map[string]bool)

	for rows.Next() {
		item := GetUsersTopResponseItem{}

		err := rows.Scan(
			&item.Id, &item.Name,
			&item.Type, &item.AuthId,
			&item.Games, &item.Metric,
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

func (s *UserAnalyticsService) getZedtimeTop(
	req GetUsersTopRequest,
) (*GetUsersTopResponse, error) {
	conds := make([]string, 0)
	args := make([]interface{}, 0)

	if req.From != nil && req.To != nil {
		conds = append(conds, "DATE(session.started_at) BETWEEN ? AND ?")
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	}

	sql := fmt.Sprintf(`
		SELECT
			users.id as user_id,
			users.name as name,
			users.auth_type,
			users.auth_id,
			total_games,
			get_avg_zt(users.id, ?, ?) as metric
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
		INNER JOIN users ON users.id = t.user_id
		ORDER BY metric DESC
		LIMIT 50`, strings.Join(conds, " AND "),
	)

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	res := GetUsersTopResponse{
		Items: []*GetUsersTopResponseItem{},
	}
	steamIdSet := make(map[string]bool)

	for rows.Next() {
		item := GetUsersTopResponseItem{}

		err := rows.Scan(
			&item.Id, &item.Name,
			&item.Type, &item.AuthId,
			&item.Games, &item.Metric,
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
