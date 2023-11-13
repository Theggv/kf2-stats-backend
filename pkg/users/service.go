package users

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/steamapi"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
)

type UserService struct {
	db *sql.DB

	steamApiService *steamapi.SteamApiUserService
}

func NewUserService(db *sql.DB) *UserService {
	service := UserService{
		db: db,
	}

	return &service
}

func (s *UserService) Inject(
	steamApiService *steamapi.SteamApiUserService,
) {
	s.steamApiService = steamApiService
}

func (s *UserService) FindCreateFind(req CreateUserRequest) (int, error) {
	data, err := s.getByAuth(req.AuthId, req.AuthType)
	if err == nil {
		return data.Id, nil
	}

	_, err = s.db.Exec(`
		INSERT INTO users (auth_id, auth_type, name) 
		VALUES (?, ?, ?)`,
		req.AuthId, req.AuthType, req.Name,
	)

	if err != nil {
		return 0, err
	}

	data, err = s.getByAuth(req.AuthId, req.AuthType)
	if err != nil {
		return 0, err
	}

	_, err = s.db.Exec(`
		INSERT INTO users_activity (user_id, current_session_id, last_session_id) 
		VALUES (?, NULL, NULL)`, data.Id,
	)

	return data.Id, err
}

func (s *UserService) GetById(id int) (*User, error) {
	row := s.db.QueryRow(`
		SELECT id, auth_id, auth_type, name FROM users WHERE id = ?`, id,
	)

	item := User{}

	err := row.Scan(&item.Id, &item.AuthId, &item.Type, &item.Name)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *UserService) getByAuth(authId string, authType models.AuthType) (*User, error) {
	row := s.db.QueryRow(`
		SELECT id, auth_id, auth_type, name FROM users WHERE auth_id = ? AND auth_type = ?`,
		authId, authType,
	)

	item := User{}

	err := row.Scan(&item.Id, &item.AuthId, &item.Type, &item.Name)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *UserService) getUserDetailed(id int) (*FilterUsersResponseUser, error) {
	sql := fmt.Sprintf(`
		SELECT 
			users.id,
			users.name,
			users.auth_id,
			users.auth_type,
			users_activity.last_session_id,
			users_activity.current_session_id,
			users_activity.updated_at
		FROM users
		INNER JOIN users_activity ON users_activity.user_id = users.id
		WHERE users.id = %v
		`, id,
	)

	item := FilterUsersResponseUser{}
	err := s.db.QueryRow(sql).Scan(
		&item.Id, &item.Name, &item.AuthId, &item.Type,
		&item.LastSessionId, &item.CurrentSessionId, &item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if item.Type == models.Steam {
		steamData, err := s.GetSteamData([]string{item.AuthId})
		if err == nil {
			if data, ok := steamData[item.AuthId]; ok {
				item.Avatar = &data.Avatar
				item.ProfileUrl = &data.ProfileUrl
			}
		}
	}

	sessionIdSet := make(map[int]bool)
	if item.LastSessionId != nil {
		sessionIdSet[*item.LastSessionId] = true
	}
	if item.CurrentSessionId != nil {
		sessionIdSet[*item.CurrentSessionId] = true
	}

	sessionIds := []int{}
	for key := range sessionIdSet {
		sessionIds = append(sessionIds, key)
	}

	if len(sessionIds) > 0 {
		sessions, err := s.getSessions(sessionIds)
		if err != nil {
			return nil, err
		}

		if item.LastSessionId != nil {
			if data, ok := sessions[*item.LastSessionId]; ok {
				item.LastSession = &data
			}
		}

		if item.CurrentSessionId != nil {
			if data, ok := sessions[*item.CurrentSessionId]; ok {
				item.CurrentSession = &data
			}
		}
	}

	return &item, nil
}

func (s *UserService) filter(req FilterUsersRequest) (*FilterUsersResponse, error) {
	page, limit := util.ParsePagination(req.Pager)

	sql := fmt.Sprintf(`
		SELECT 
			users.id,
			users.name,
			users.auth_id,
			users.auth_type,
			users_activity.last_session_id,
			users_activity.current_session_id,
			users_activity.updated_at
		FROM users
		INNER JOIN users_activity ON users_activity.user_id = users.id
		ORDER BY users_activity.updated_at DESC
		LIMIT %v, %v
		`, page*limit, limit,
	)

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	sessionIdSet := make(map[int]bool)
	steamIdSet := make(map[string]bool)
	items := []*FilterUsersResponseUser{}

	for rows.Next() {
		item := FilterUsersResponseUser{}

		rows.Scan(
			&item.Id, &item.Name, &item.AuthId, &item.Type,
			&item.LastSessionId, &item.CurrentSessionId, &item.UpdatedAt,
		)

		if item.LastSessionId != nil {
			sessionIdSet[*item.LastSessionId] = true
		}
		if item.CurrentSessionId != nil {
			sessionIdSet[*item.CurrentSessionId] = true
		}
		if item.Type == models.Steam {
			steamIdSet[item.AuthId] = true
		}

		items = append(items, &item)
	}

	{
		steamIds := []string{}
		for key := range steamIdSet {
			steamIds = append(steamIds, key)
		}

		steamData, err := s.GetSteamData(steamIds)
		if err != nil {
			return nil, err
		}

		for _, item := range items {
			if data, ok := steamData[item.AuthId]; ok {
				item.Avatar = &data.Avatar
				item.ProfileUrl = &data.ProfileUrl
			}
		}
	}

	sessionIds := []int{}
	for key := range sessionIdSet {
		sessionIds = append(sessionIds, key)
	}

	if len(sessionIds) > 0 {
		sessions, err := s.getSessions(sessionIds)
		if err != nil {
			return nil, err
		}

		for _, item := range items {
			if item.LastSessionId != nil {
				if data, ok := sessions[*item.LastSessionId]; ok {
					item.LastSession = &data
				}
			}

			if item.CurrentSessionId != nil {
				if data, ok := sessions[*item.CurrentSessionId]; ok {
					item.CurrentSession = &data
				}
			}
		}
	}

	return &FilterUsersResponse{
		Items: items,
	}, nil
}

func (s *UserService) getSessions(sessionIds []int) (map[int]FilterUsersResponseUserSession, error) {
	sql := fmt.Sprintf(`
		SELECT 
			session.id,
			session.mode,
			session.length,
			session.diff,
			session.status,
			gd.wave,
			cd.spawn_cycle,
			cd.max_monsters,
			cd.wave_size_fakes,
			cd.zeds_type,
			server.name,
			maps.name
		FROM session
		INNER JOIN server ON server.id = session.server_id
		INNER JOIN maps ON maps.id = session.map_id
		INNER JOIN session_game_data gd ON gd.session_id = session.id
		LEFT JOIN session_game_data_cd cd ON cd.session_id = session.id
		WHERE session.id IN (%v)`,
		util.IntArrayToString(sessionIds, ","),
	)

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	items := make(map[int]FilterUsersResponseUserSession)
	for rows.Next() {
		item := FilterUsersResponseUserSession{}
		cdData := models.CDGameData{}

		rows.Scan(
			&item.Id, &item.Mode, &item.Length, &item.Difficulty, &item.Status, &item.Wave,
			&cdData.SpawnCycle, &cdData.MaxMonsters, &cdData.WaveSizeFakes, &cdData.ZedsType,
			&item.ServerName, &item.MapName,
		)

		if cdData.SpawnCycle != nil {
			item.CDData = &cdData
		}

		items[item.Id] = item
	}

	return items, nil
}

func (s *UserService) GetSteamData(steamIds []string) (map[string]steamapi.GetUserSummaryPlayer, error) {
	steamData, err := s.steamApiService.GetUserSummary(steamIds)
	if err != nil {
		return nil, err
	}

	steamDataSet := make(map[string]steamapi.GetUserSummaryPlayer)
	for _, data := range steamData {
		steamDataSet[data.SteamId] = data
	}

	return steamDataSet, nil
}

func (s *UserService) getRecentSessions(req RecentSessionsRequest) (*RecentSessionsResponse, error) {
	page, limit := util.ParsePagination(req.Pager)

	perkConds := []string{}
	perkConds = append(perkConds, fmt.Sprintf("wsp.player_id = %v", req.UserId))

	sql := fmt.Sprintf(`
		SELECT
			session.id,
			session.mode,
			session.length,
			session.diff,
			session.status,
			gd.wave,
			cd.spawn_cycle,
			cd.max_monsters,
			cd.wave_size_fakes,
			cd.zeds_type,
			maps.name,
			server.id,
			server.name,
			max(wsp.created_at)
		FROM session
		INNER JOIN wave_stats ws ON ws.session_id = session.id
		INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
		INNER JOIN maps on maps.id = session.map_id
		INNER JOIN server on server.id = session.server_id
		INNER JOIN session_game_data gd on gd.session_id = session.id
		LEFT JOIN session_game_data_cd cd on cd.session_id = session.id
		WHERE wsp.player_id = %v
		GROUP BY session.id
		ORDER BY session.id DESC
		LIMIT %v, %v
		`, req.UserId, page*limit, limit,
	)

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	sessionMap := make(map[int]*RecentSessionsResponseSession)
	items := []*RecentSessionsResponseSession{}
	for rows.Next() {
		item := RecentSessionsResponseSession{}
		cdData := models.CDGameData{}

		rows.Scan(
			&item.Id, &item.Mode, &item.Length, &item.Difficulty, &item.Status, &item.Wave,
			&cdData.SpawnCycle, &cdData.MaxMonsters, &cdData.WaveSizeFakes, &cdData.ZedsType,
			&item.MapName, &item.Server.Id, &item.Server.Name, &item.UpdatedAt,
		)

		if cdData.SpawnCycle != nil {
			item.CDData = &cdData
		}

		items = append(items, &item)
		sessionMap[item.Id] = &item
	}

	{
		ids := []int{}
		for id := range sessionMap {
			ids = append(ids, id)
		}
		perkConds = append(perkConds, fmt.Sprintf("session.id IN (%v)", util.IntArrayToString(ids, ",")))

		sql = fmt.Sprintf(`
			SELECT DISTINCT session.id, wsp.perk
			FROM session
			INNER JOIN wave_stats ws ON ws.session_id = session.id
			INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
			INNER JOIN maps on maps.id = session.map_id
			WHERE %v
			`, strings.Join(perkConds, " AND "),
		)

		rows, err = s.db.Query(sql)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var sessionId, perk int

			rows.Scan(&sessionId, &perk)

			if data, ok := sessionMap[sessionId]; ok {
				data.Perks = append(data.Perks, perk)
			}
		}
	}

	return &RecentSessionsResponse{
		Items: items,
	}, nil
}
