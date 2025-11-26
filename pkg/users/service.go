package users

import (
	"database/sql"
	"fmt"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/steamapi"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
	"github.com/theggv/kf2-stats-backend/pkg/session/difficulty"
)

type UserService struct {
	db *sql.DB

	steamApiService *steamapi.SteamApiUserService
	diffService     *difficulty.DifficultyCalculatorService
}

func NewUserService(db *sql.DB) *UserService {
	service := UserService{
		db: db,
	}

	return &service
}

func (s *UserService) Inject(
	steamApiService *steamapi.SteamApiUserService,
	diffService *difficulty.DifficultyCalculatorService,
) {
	s.steamApiService = steamApiService
	s.diffService = diffService
}

func (s *UserService) FindCreateFind(req CreateUserRequest) (int, error) {
	if data, err := s.GetByAuth(req.AuthId, req.AuthType); err == nil {
		_, err = s.db.Exec(`
			UPDATE users SET name = ? WHERE id = ?`,
			req.Name, data.Id,
		)
		if err != nil {
			return 0, err
		}

		return data.Id, nil
	}

	_, err := s.db.Exec(`
		INSERT INTO users (auth_id, auth_type, name) 
		VALUES (?, ?, ?)`,
		req.AuthId, req.AuthType, req.Name,
	)

	if err != nil {
		return 0, err
	}

	data, err := s.GetByAuth(req.AuthId, req.AuthType)
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

func (s *UserService) GetManyById(userId []int) ([]*User, error) {
	if len(userId) == 0 {
		return []*User{}, nil
	}

	sql := fmt.Sprintf(`
		SELECT id, auth_id, auth_type, name FROM users WHERE id IN (%v)`,
		util.IntArrayToString(userId, ","),
	)

	rows, err := s.db.Query(sql)

	if err != nil {
		return nil, err
	}

	items := []*User{}
	for rows.Next() {
		item := User{}

		err := rows.Scan(&item.Id, &item.AuthId, &item.Type, &item.Name)
		if err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	return items, nil
}

func (s *UserService) GetByAuth(authId string, authType models.AuthType) (*User, error) {
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
				item.LastSession = data
			}
		}

		if item.CurrentSessionId != nil {
			if data, ok := sessions[*item.CurrentSessionId]; ok {
				item.CurrentSession = data
			}
		}
	}

	return &item, nil
}

func (s *UserService) filter(req FilterUsersRequest) (*FilterUsersResponse, error) {
	page, limit := util.ParsePagination(req.Pager)

	stmt := fmt.Sprintf(`
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
		WHERE LOWER(users.name) LIKE ?
		ORDER BY users_activity.updated_at DESC
		LIMIT %v, %v
		`, page*limit, limit,
	)
	args := []any{
		fmt.Sprintf("%%%v%%", req.SearchText),
	}

	rows, err := s.db.Query(stmt, args...)
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
					item.LastSession = data
				}
			}

			if item.CurrentSessionId != nil {
				if data, ok := sessions[*item.CurrentSessionId]; ok {
					item.CurrentSession = data
				}
			}
		}
	}

	var total int
	{
		// Prepare count query
		stmt = `
			SELECT count(*) FROM users
			INNER JOIN users_activity ON users_activity.user_id = users.id
			WHERE lower(users.name) LIKE ?`

		args := []any{
			fmt.Sprintf("%%%v%%", req.SearchText),
		}

		if err := s.db.QueryRow(stmt, args...).Scan(&total); err != nil {
			return nil, err
		}
	}

	return &FilterUsersResponse{
		Items: items,
		Metadata: models.PaginationResponse{
			Page:           page,
			ResultsPerPage: limit,
			TotalResults:   total,
		},
	}, nil
}

func (s *UserService) getSessions(sessionIds []int) (map[int]*FilterUsersResponseUserSession, error) {
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
		LEFT JOIN session_game_data_extra cd ON cd.session_id = session.id
		WHERE session.id IN (%v)`,
		util.IntArrayToString(sessionIds, ","),
	)

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	items := map[int]*FilterUsersResponseUserSession{}
	for rows.Next() {
		item := FilterUsersResponseUserSession{}
		cdData := models.ExtraGameData{}

		rows.Scan(
			&item.Id, &item.Mode, &item.Length, &item.Difficulty, &item.Status, &item.Wave,
			&cdData.SpawnCycle, &cdData.MaxMonsters, &cdData.WaveSizeFakes, &cdData.ZedsType,
			&item.ServerName, &item.MapName,
		)

		if cdData.SpawnCycle != nil {
			item.CDData = &cdData
		}

		items[item.Id] = &item
	}

	// Join session difficulty
	{
		difficulty, err := s.diffService.GetByIds(sessionIds, false)
		if err != nil {
			return nil, err
		}

		for i := range items {
			for j := range difficulty {
				if items[i].Id == difficulty[j].SessionId {
					items[i].Metadata.Difficulty = difficulty[j]
				}
			}
		}
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

func (s *UserService) GetUserProfiles(
	userId []int,
) ([]*models.UserProfile, error) {
	users, err := s.GetManyById(userId)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return []*models.UserProfile{}, nil
	}

	set := make(map[string]bool)
	steamIds := []string{}

	for _, player := range users {
		if player.Type != models.Steam {
			continue
		}
		set[player.AuthId] = true
	}

	for key := range set {
		steamIds = append(steamIds, key)
	}

	steamData, err := s.steamApiService.GetUserSummary(steamIds)
	if err != nil {
		return nil, err
	}

	steamDataSet := make(map[string]steamapi.GetUserSummaryPlayer)
	for _, data := range steamData {
		steamDataSet[data.SteamId] = data
	}

	profiles := []*models.UserProfile{}

	for _, player := range users {
		profile := models.UserProfile{
			Id:     player.Id,
			AuthId: player.AuthId,
			Name:   player.Name,
		}

		if data, exists := steamDataSet[player.AuthId]; exists {
			profile.ProfileUrl = &data.ProfileUrl
			profile.Avatar = &data.Avatar
		}

		profiles = append(profiles, &profile)
	}

	return profiles, nil
}
