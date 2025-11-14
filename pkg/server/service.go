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

type userIdTuple struct {
	SessionId         int
	WaveStatsPlayerId int
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
			wsp.created_at,
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

	waveStatsPlayerIdSet := make(map[int]userIdTuple)
	steamIdSet := make(map[string]bool)
	items := []*RecentUsersResponseUser{}

	var total int
	for rows.Next() {
		item := RecentUsersResponseUser{}

		rows.Scan(
			&item.Id, &item.Name, &item.AuthId, &item.Type,
			&item.SessionId, &item.WaveStatsPlayerId, &item.UpdatedAt,
			&total,
		)

		waveStatsPlayerIdSet[item.Id] = userIdTuple{
			SessionId:         item.SessionId,
			WaveStatsPlayerId: item.WaveStatsPlayerId,
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

		steamData, err := s.userService.GetSteamData(steamIds)
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

	{
		sessions, err := s.getSessions(waveStatsPlayerIdSet)
		if err != nil {
			return nil, err
		}

		for _, item := range items {
			if data, ok := sessions[item.Id]; ok {
				item.Session = data
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
	userIds map[int]userIdTuple,
) (map[int]*RecentUsersResponseUserSession, error) {
	wspIds := []int{}
	for _, value := range userIds {
		wspIds = append(wspIds, value.WaveStatsPlayerId)
	}

	stmt := fmt.Sprintf(`
		SELECT
			wsp.player_id,
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
			maps.name
		FROM session
		INNER JOIN wave_stats ws ON ws.session_id = session.id
		INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
		INNER JOIN maps ON maps.id = session.map_id
		INNER JOIN session_game_data gd ON gd.session_id = session.id
		LEFT JOIN session_game_data_extra cd ON cd.session_id = session.id
		WHERE wsp.id IN (%v)
		`, util.IntArrayToString(wspIds, ","),
	)

	rows, err := s.db.Query(stmt)
	if err != nil {
		return nil, err
	}

	perkConds := []string{}
	items := map[int]*RecentUsersResponseUserSession{}
	lookup := map[int]bool{}
	for rows.Next() {
		item := RecentUsersResponseUserSession{}
		cdData := models.ExtraGameData{}

		rows.Scan(
			&item.PlayerId, &item.Id, &item.Mode, &item.Length, &item.Difficulty, &item.Status, &item.Wave,
			&cdData.SpawnCycle, &cdData.MaxMonsters, &cdData.WaveSizeFakes, &cdData.ZedsType,
			&item.MapName,
		)

		if cdData.SpawnCycle != nil {
			item.CDData = &cdData
		}

		items[item.PlayerId] = &item
		lookup[item.Id] = true
		perkConds = append(perkConds,
			fmt.Sprintf("(session.id = %v AND wsp.player_id = %v)", item.Id, item.PlayerId),
		)
	}

	// Join perks
	{
		stmt = fmt.Sprintf(`
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
			var playerId, perk int

			rows.Scan(&playerId, &perk)

			if data, ok := items[playerId]; ok {
				data.Perks = append(data.Perks, perk)
			}
		}
	}

	// Join session difficulty
	{
		sessionIds := []int{}
		for sessionId := range lookup {
			sessionIds = append(sessionIds, sessionId)
		}

		difficulty, err := s.diffService.GetByIds(sessionIds)
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
