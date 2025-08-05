package matches

import (
	"fmt"
	"strings"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
)

type filterServerData struct {
	MatchId int
	Server  *MatchServer
}

func (s *MatchesService) getServerData(matchId []int) ([]*filterServerData, error) {
	if len(matchId) == 0 {
		return []*filterServerData{}, nil
	}

	sql := fmt.Sprintf(`
		SELECT session.id, server.name, server.address FROM session
		INNER JOIN server ON server.id = session.server_id
		WHERE %v`,
		fmt.Sprintf("session.id in (%v)", util.IntArrayToString(matchId, ",")),
	)

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	items := []*filterServerData{}

	for rows.Next() {
		server := MatchServer{}
		item := filterServerData{}

		rows.Scan(&item.MatchId, &server.Name, &server.Address)
		item.Server = &server

		items = append(items, &item)
	}

	return items, nil
}

type filterMapData struct {
	MatchId int
	Map     *MatchMap
}

func (s *MatchesService) getMapData(matchId []int) ([]*filterMapData, error) {
	if len(matchId) == 0 {
		return []*filterMapData{}, nil
	}

	sql := fmt.Sprintf(`
		SELECT session.id, maps.name, maps.preview FROM session
		INNER JOIN maps ON maps.id = session.map_id
		WHERE %v`,
		fmt.Sprintf("session.id in (%v)", util.IntArrayToString(matchId, ",")),
	)

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	items := []*filterMapData{}

	for rows.Next() {
		mapData := MatchMap{}
		item := filterMapData{}

		rows.Scan(&item.MatchId, &mapData.Name, &mapData.Preview)
		item.Map = &mapData

		items = append(items, &item)
	}

	return items, nil
}

type filterGameData struct {
	MatchId  int
	GameData *models.GameData
}

func (s *MatchesService) getGameData(matchId []int) ([]*filterGameData, error) {
	if len(matchId) == 0 {
		return []*filterGameData{}, nil
	}

	sql := fmt.Sprintf(`
		SELECT session.id, 
			gd.max_players, gd.players_online, gd.players_alive,
			gd.wave, gd.is_trader_time, gd.zeds_left 
		FROM session
		INNER JOIN session_game_data gd ON session.id = gd.session_id
		WHERE %v`,
		fmt.Sprintf("session.id in (%v)", util.IntArrayToString(matchId, ",")),
	)

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	items := []*filterGameData{}

	for rows.Next() {
		gameData := models.GameData{}
		item := filterGameData{}

		rows.Scan(&item.MatchId,
			&gameData.MaxPlayers, &gameData.PlayersOnline, &gameData.PlayersAlive,
			&gameData.Wave, &gameData.IsTraderTime, &gameData.ZedsLeft,
		)

		item.GameData = &gameData
		items = append(items, &item)
	}

	return items, nil
}

type filterCDGameData struct {
	MatchId int
	CDData  *models.ExtraGameData
}

func (s *MatchesService) getCDGameData(matchId []int) ([]*filterCDGameData, error) {
	if len(matchId) == 0 {
		return []*filterCDGameData{}, nil
	}

	sql := fmt.Sprintf(`
		SELECT session.id, 
			cd.spawn_cycle, cd.max_monsters,
			cd.wave_size_fakes, cd.zeds_type
		FROM session
		INNER JOIN session_game_data_extra cd ON session.id = cd.session_id
		WHERE %v`,
		fmt.Sprintf("session.id in (%v)", util.IntArrayToString(matchId, ",")),
	)

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	items := []*filterCDGameData{}

	for rows.Next() {
		cdData := models.ExtraGameData{}
		item := filterCDGameData{}

		rows.Scan(&item.MatchId,
			&cdData.SpawnCycle, &cdData.MaxMonsters,
			&cdData.WaveSizeFakes, &cdData.ZedsType,
		)

		item.CDData = &cdData
		items = append(items, &item)
	}

	return items, nil
}

type filterPlayerData struct {
	MatchId    int
	Players    []*MatchPlayer
	Spectators []*MatchPlayer
}

func (s *MatchesService) getPlayerData(matchId []int) ([]*filterPlayerData, error) {
	if len(matchId) == 0 {
		return []*filterPlayerData{}, nil
	}

	sql := fmt.Sprintf(`
		SELECT
			current_session_id, user_id,
			perk, level, prestige,
			health, armor, is_spectator
		FROM users_activity activity
		WHERE %v`,
		fmt.Sprintf("current_session_id in (%v)", util.IntArrayToString(matchId, ",")),
	)

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	type playerData struct {
		UserId      int
		MatchId     int
		IsSpectator bool
		Player      *MatchPlayer
	}

	players := []*playerData{}
	userIdSet := make(map[int]bool)

	defer rows.Close()
	for rows.Next() {
		player := playerData{}
		matchData := MatchPlayer{}

		rows.Scan(&player.MatchId, &player.UserId,
			&matchData.Perk, &matchData.Level, &matchData.Prestige,
			&matchData.Health, &matchData.Armor, &player.IsSpectator,
		)

		userIdSet[player.UserId] = true

		player.Player = &matchData
		players = append(players, &player)
	}

	{
		// Join profiles
		userId := []int{}
		for key := range userIdSet {
			userId = append(userId, key)
		}

		profiles, err := s.userService.GetUserProfiles(userId)
		if err != nil {
			return nil, err
		}

		for i := range players {
			for j := range profiles {
				if players[i].UserId == profiles[j].Id {
					players[i].Player.Profile = profiles[j]
				}
			}
		}
	}

	// Group players by matchId
	items := []*filterPlayerData{}
	itemsMap := make(map[int]*filterPlayerData)

	for i := range players {
		item := players[i]
		val, ok := itemsMap[item.MatchId]

		if !ok {
			val = &filterPlayerData{
				MatchId:    item.MatchId,
				Players:    []*MatchPlayer{},
				Spectators: []*MatchPlayer{},
			}

			itemsMap[item.MatchId] = val
		}

		if item.IsSpectator {
			val.Spectators = append(val.Spectators, item.Player)
		} else {
			val.Players = append(val.Players, item.Player)
		}
	}

	for _, value := range itemsMap {
		items = append(items, value)
	}

	return items, nil
}

func (s *MatchesService) Filter(req FilterMatchesRequest) (*FilterMatchesResponse, error) {
	page, limit := util.ParsePagination(req.Pager)

	attributes := []string{}
	conditions := []string{}
	joins := []string{}
	order := "asc"

	// Prepare fields
	attributes = append(attributes,
		"session.id", "session.server_id", "session.map_id",
		"session.mode", "session.length", "session.diff",
		"session.status", "session.created_at", "session.updated_at",
		"session.started_at", "session.completed_at",
	)

	// Prepare filter query
	conditions = append(conditions, "1") // in case if no filters passed

	if len(req.ServerId) > 0 {
		conditions = append(conditions,
			fmt.Sprintf("session.server_id in (%s)", util.IntArrayToString(req.ServerId, ",")),
		)
	}

	if len(req.MapId) > 0 {
		conditions = append(conditions,
			fmt.Sprintf("session.map_id in (%s)", util.IntArrayToString(req.MapId, ",")),
		)
	}

	if req.Difficulty != nil {
		conditions = append(conditions, fmt.Sprintf("session.diff = %v", *req.Difficulty))
	}

	if req.Length != nil {
		if *req.Length == models.Custom {
			conditions = append(conditions, fmt.Sprintf("session.length NOT IN (%v, %v, %v)",
				models.Short, models.Medium, models.Long))
		} else {
			conditions = append(conditions, fmt.Sprintf("session.length = %v", *req.Length))
		}
	}

	if req.Mode != nil {
		conditions = append(conditions, fmt.Sprintf("session.mode = %v", *req.Mode))
	}

	if len(req.Status) > 0 {
		conditions = append(conditions,
			fmt.Sprintf("session.status in (%s)", util.IntArrayToString(req.Status, ",")),
		)
	}

	// Order
	if req.ReverseOrder != nil && *req.ReverseOrder {
		order = "desc"
	}

	sql := fmt.Sprintf(`
		SELECT %v FROM session
		%v
		WHERE %v
		ORDER BY session.updated_at %v
		LIMIT %v, %v`,
		strings.Join(attributes, ", "),
		strings.Join(joins, "\n"),
		strings.Join(conditions, " AND "), order, page*limit, limit,
	)

	// Execute filter query
	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	matchId := []int{}
	items := []*Match{}

	defer rows.Close()
	for rows.Next() {
		item := Match{}
		sessionData := MatchSession{}

		fields := []any{
			&sessionData.SessionId, &sessionData.ServerId, &sessionData.MapId,
			&sessionData.Mode, &sessionData.Length, &sessionData.Difficulty,
			&sessionData.Status, &sessionData.CreatedAt, &sessionData.UpdatedAt,
			&sessionData.StartedAt, &sessionData.CompletedAt,
		}

		err := rows.Scan(fields...)
		if err != nil {
			fmt.Print(err)
			continue
		}

		item.Session = sessionData

		matchId = append(matchId, item.Session.SessionId)
		items = append(items, &item)
	}

	if req.IncludeServer != nil && *req.IncludeServer {
		serverData, err := s.getServerData(matchId)
		if err != nil {
			return nil, err
		}

		for i := range items {
			for j := range serverData {
				if items[i].Session.SessionId == serverData[j].MatchId {
					items[i].Server = serverData[j].Server
				}
			}
		}
	}

	if req.IncludeMap != nil && *req.IncludeMap {
		mapData, err := s.getMapData(matchId)
		if err != nil {
			return nil, err
		}

		for i := range items {
			for j := range mapData {
				if items[i].Session.SessionId == mapData[j].MatchId {
					items[i].Map = mapData[j].Map
				}
			}
		}
	}

	if req.IncludeGameData != nil && *req.IncludeGameData {
		gameData, err := s.getGameData(matchId)
		if err != nil {
			return nil, err
		}

		for i := range items {
			for j := range gameData {
				if items[i].Session.SessionId == gameData[j].MatchId {
					items[i].GameData = gameData[j].GameData
				}
			}
		}
	}

	if req.IncludeCDData != nil && *req.IncludeCDData {
		cdData, err := s.getCDGameData(matchId)
		if err != nil {
			return nil, err
		}

		for i := range items {
			for j := range cdData {
				if items[i].Session.SessionId == cdData[j].MatchId {
					items[i].CDData = cdData[j].CDData
				}
			}
		}
	}

	if req.IncludePlayers != nil && *req.IncludePlayers {
		playerData, err := s.getPlayerData(matchId)
		if err != nil {
			return nil, err
		}

		for i := range items {
			for j := range playerData {
				if items[i].Session.SessionId == playerData[j].MatchId {
					items[i].Players = playerData[j].Players
					items[i].Spectators = playerData[j].Spectators
				}
			}
		}
	}

	var total int
	{
		sql = fmt.Sprintf(`
			SELECT count(*) FROM session
			%v
			WHERE %v`,
			strings.Join(joins, "\n"),
			strings.Join(conditions, " AND "),
		)

		if err := s.db.QueryRow(sql).Scan(&total); err != nil {
			return nil, err
		}
	}

	return &FilterMatchesResponse{
		Items: items,
		Metadata: models.PaginationResponse{
			Page:           page,
			ResultsPerPage: limit,
			TotalResults:   total,
		},
	}, nil
}
