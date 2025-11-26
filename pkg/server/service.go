package server

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
	"github.com/theggv/kf2-stats-backend/pkg/session/difficulty"
	"github.com/theggv/kf2-stats-backend/pkg/users"
)

type ServerService struct {
	db *sql.DB

	userService *users.UserService
	diffService *difficulty.DifficultyCalculatorService
}

func (s *ServerService) Inject(
	userService *users.UserService,
	diffService *difficulty.DifficultyCalculatorService,
) {
	s.userService = userService
	s.diffService = diffService
}

func NewServerService(db *sql.DB) *ServerService {
	service := ServerService{
		db: db,
	}

	return &service
}

func (s *ServerService) Create(req AddServerRequest) (int, error) {
	_, err := s.db.Exec(`
		INSERT INTO server (name, address) VALUES (?, ?)
			ON DUPLICATE KEY UPDATE name = ?`,
		req.Name, req.Address, req.Name)

	if err != nil {
		return 0, err
	}

	data, err := s.getByAddress(req.Address)
	if err != nil {
		return 0, err
	}

	return data.Id, err
}

func (s *ServerService) GetByPattern(pattern string) ([]Server, error) {
	sqlPattern := "%" + pattern + "%"
	rows, err := s.db.Query(`
		SELECT id, name, address FROM server 
		WHERE (address LIKE ?) OR (name LIKE ?)
		ORDER BY name`, sqlPattern, sqlPattern)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	results := []Server{}

	for rows.Next() {
		server := Server{}

		err := rows.Scan(&server.Id, &server.Name, &server.Address)
		if err != nil {
			continue
		}

		results = append(results, server)
	}

	return results, nil
}

func (s *ServerService) GetById(id int) (*Server, error) {
	row := s.db.QueryRow(`SELECT id, name, address FROM server WHERE id = ?`, id)

	server := Server{}

	err := row.Scan(&server.Id, &server.Name, &server.Address)
	if err != nil {
		return nil, err
	}

	return &server, nil
}

func (s *ServerService) getByAddress(address string) (*Server, error) {
	row := s.db.QueryRow(`SELECT id, name, address FROM server WHERE address = ?`, address)

	server := Server{}

	err := row.Scan(&server.Id, &server.Name, &server.Address)
	if err != nil {
		return nil, err
	}

	return &server, nil
}

func (s *ServerService) UpdateName(data UpdateNameRequest) error {
	_, err := s.db.Exec(`UPDATE server SET name = ? WHERE id = ?`,
		data.Name, data.Id)

	return err
}

func (s *ServerService) GetRecentUsers(req RecentUsersRequest) (*RecentUsersResponse, error) {
	page, limit := util.ParsePagination(req.Pager)

	conds := []string{}
	conds = append(conds, fmt.Sprintf("session.server_id = %v", req.ServerId))

	sql := fmt.Sprintf(`
		WITH ranking AS (
			SELECT 
				wsp.player_id user_id,
				wsp.id wsp_id,
				DENSE_RANK() OVER w as rating
			FROM session
			INNER JOIN wave_stats ws ON ws.session_id = session.id
			INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
			WHERE %v
			WINDOW w AS (PARTITION BY wsp.player_id ORDER BY wsp.id DESC)
			ORDER BY wsp.id DESC
		), recent_players AS (
			SELECT user_id, wsp_id
			FROM ranking cte
			WHERE rating = 1
		), pagination AS (
			SELECT user_id, wsp_id 
			FROM recent_players cte
			LIMIT %v, %v
		), metadata AS (
			SELECT count(distinct user_id) as total 
			FROM recent_players cte
		)
		SELECT 
			users.id,
			users.name,
			users.auth_id,
			users.auth_type,
			ws.session_id,
			wsp.id wsp_id,
			metadata.total as total_players
		FROM pagination cte
		INNER JOIN wave_stats_player wsp ON wsp.id = cte.wsp_id
		INNER JOIN wave_stats ws ON ws.id = wsp.stats_id
		INNER JOIN users ON users.id = wsp.player_id
		CROSS JOIN metadata
		`, strings.Join(conds, " AND "), page*limit, limit,
	)

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	wspIds := []int{}
	steamIdSet := make(map[string]bool)
	items := []*RecentUsersResponseUser{}

	var total int
	for rows.Next() {
		item := RecentUsersResponseUser{}
		profile := models.UserProfile{}

		fields := []any{
			&profile.Id, &profile.Name, &profile.AuthId, &profile.Type,
			&item.SessionId, &item.WaveStatsPlayerId,
			&total,
		}

		rows.Scan(fields...)

		wspIds = append(wspIds, item.WaveStatsPlayerId)

		if profile.Type == models.Steam {
			steamIdSet[profile.AuthId] = true
		}

		item.UserProfile = &profile

		items = append(items, &item)
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

		for _, item := range items {
			if data, ok := steamData[item.UserProfile.AuthId]; ok {
				item.UserProfile.Avatar = &data.Avatar
				item.UserProfile.ProfileUrl = &data.ProfileUrl
			}
		}
	}

	{
		sessions, err := s.getSessions(wspIds)
		if err != nil {
			return nil, err
		}

		for _, item := range items {
			if data, ok := sessions[item.UserProfile.Id]; ok {
				item.Match = data
			}
		}
	}

	return &RecentUsersResponse{
		Items: items,
		Metadata: models.PaginationResponse{
			Page:           page,
			ResultsPerPage: limit,
			TotalResults:   total,
		},
	}, nil
}

func (s *ServerService) getSessions(
	wspIds []int,
) (map[int]*models.Match, error) {
	fields := []string{}

	fields = append(fields,
		"wsp.player_id", "wsp.created_at",
		"session.id", "session.server_id", "session.map_id",
		"session.mode", "session.length", "session.diff",
		"session.status", "session.created_at", "session.updated_at",
		"session.started_at", "session.completed_at",
		"maps.name", "gd.wave",
		"extra.spawn_cycle", "extra.max_monsters", "extra.wave_size_fakes", "extra.zeds_type",
	)

	stmt := fmt.Sprintf(`
		SELECT
		%v
		FROM session
		INNER JOIN wave_stats ws ON ws.session_id = session.id
		INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
		INNER JOIN maps ON maps.id = session.map_id
		INNER JOIN session_game_data gd ON gd.session_id = session.id
		LEFT JOIN session_game_data_extra extra ON extra.session_id = session.id
		WHERE wsp.id IN (%v)
		`, strings.Join(fields, ", "), util.IntArrayToString(wspIds, ","),
	)

	rows, err := s.db.Query(stmt)
	if err != nil {
		return nil, err
	}

	perkConds := []string{}
	items := map[int]*models.Match{}
	lookup := map[int]bool{}

	for rows.Next() {
		var userId int
		sessionData := models.MatchSession{}
		gameData := models.GameData{}
		extraData := models.ExtraGameData{}
		mapData := models.MatchMap{}
		userData := models.MatchUserData{}

		fields := []any{
			&userId, &userData.LastSeen,
			&sessionData.Id, &sessionData.ServerId, &sessionData.MapId,
			&sessionData.Mode, &sessionData.Length, &sessionData.Difficulty,
			&sessionData.Status, &sessionData.CreatedAt, &sessionData.UpdatedAt,
			&sessionData.StartedAt, &sessionData.CompletedAt,
			&mapData.Name, &gameData.Wave,
			&extraData.SpawnCycle, &extraData.MaxMonsters, &extraData.WaveSizeFakes, &extraData.ZedsType,
		}

		rows.Scan(fields...)

		item := models.Match{
			Session: sessionData,
			Details: models.MatchDetails{
				Map:      &mapData,
				GameData: &gameData,
				UserData: &userData,
			},
		}

		if extraData.SpawnCycle != nil {
			item.Details.ExtraGameData = &extraData
		}

		items[userId] = &item
		lookup[sessionData.Id] = true
		perkConds = append(perkConds,
			fmt.Sprintf("(session.id = %v AND wsp.player_id = %v)", sessionData.Id, userId),
		)
	}

	// Join perks
	{
		stmt := fmt.Sprintf(`
			SELECT DISTINCT wsp.player_id, perk
			FROM session
			INNER JOIN wave_stats ws ON ws.session_id = session.id
			INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
			INNER JOIN maps on maps.id = session.map_id
			WHERE %v
			`, strings.Join(perkConds, " OR "),
		)

		rows, err = s.db.Query(stmt)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var userId, perk int

			rows.Scan(&userId, &perk)

			if data, ok := items[userId]; ok {
				data.Details.UserData.Perks = append(data.Details.UserData.Perks, perk)
			}
		}
	}

	// Join session difficulty
	{
		sessionIds := []int{}
		for sessionId := range lookup {
			sessionIds = append(sessionIds, sessionId)
		}

		difficulty, err := s.diffService.GetByIds(sessionIds, false)
		if err != nil {
			return nil, err
		}

		for i := range items {
			for j := range difficulty {
				if items[i].Session.Id == difficulty[j].SessionId {
					items[i].Metadata.Difficulty = difficulty[j]
				}
			}
		}
	}

	return items, nil
}

func (s *ServerService) GetLastSession(id int) (*ServerLastSessionResponse, error) {

	res := ServerLastSessionResponse{}

	err := s.db.
		QueryRow(`
			SELECT id, status
			FROM session
			WHERE server_id = ?
			ORDER BY id DESC
			LIMIT 1`, id,
		).
		Scan(&res.Id, &res.Status)

	if err != nil {
		return nil, err
	}

	return &res, nil
}
