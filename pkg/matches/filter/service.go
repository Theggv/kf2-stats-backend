package filter

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/steamapi"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
	"github.com/theggv/kf2-stats-backend/pkg/maps"
	"github.com/theggv/kf2-stats-backend/pkg/server"
	"github.com/theggv/kf2-stats-backend/pkg/session"
	"github.com/theggv/kf2-stats-backend/pkg/session/difficulty"
	"github.com/theggv/kf2-stats-backend/pkg/users"
)

type MatchesFilterService struct {
	db *sql.DB

	userService     *users.UserService
	sessionService  *session.SessionService
	diffService     *difficulty.DifficultyCalculatorService
	mapsService     *maps.MapsService
	serverService   *server.ServerService
	steamApiService *steamapi.SteamApiUserService
}

func (s *MatchesFilterService) Inject(
	userService *users.UserService,
	sessionService *session.SessionService,
	difficultyService *difficulty.DifficultyCalculatorService,
	mapsService *maps.MapsService,
	serverService *server.ServerService,
	steamApiService *steamapi.SteamApiUserService,
) {
	s.userService = userService
	s.sessionService = sessionService
	s.diffService = difficultyService
	s.mapsService = mapsService
	s.serverService = serverService
	s.steamApiService = steamApiService
}

func NewMatchesService(db *sql.DB) *MatchesFilterService {
	service := MatchesFilterService{
		db: db,
	}

	return &service
}

type filterServerData struct {
	MatchId int
	Server  *MatchServer
}

func (s *MatchesFilterService) getServerData(matchId []int) ([]*filterServerData, error) {
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

func (s *MatchesFilterService) getMapData(matchId []int) ([]*filterMapData, error) {
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

func (s *MatchesFilterService) getGameData(matchId []int) ([]*filterGameData, error) {
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

func (s *MatchesFilterService) getExtraGameData(matchId []int) ([]*filterCDGameData, error) {
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

func (s *MatchesFilterService) getPlayerData(matchId []int) ([]*filterPlayerData, error) {
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

func (s *MatchesFilterService) Filter(req FilterMatchesRequest) (*FilterMatchesResponse, error) {
	page, limit := req.Pager.Parse()

	fields := []string{}
	conds := []string{}
	args := []any{}

	// Prepare fields
	fields = append(fields,
		"session.id", "session.server_id", "session.map_id",
		"session.mode", "session.length", "session.diff",
		"session.status", "session.created_at", "session.updated_at",
		"session.started_at", "session.completed_at",
	)

	fieldsMapper := map[string]string{
		"updated_at":           "updated_at",
		"diff_final_score":     "final_score",
		"diff_potential_score": "potential_score",
	}
	sortBy, direction := req.SortBy.Transform(fieldsMapper, "updated_at")

	// Prepare filter query
	conds = append(conds, "1") // in case if no filters passed

	if req.From != nil && req.To != nil {
		conds = append(conds, "DATE(session.updated_at) BETWEEN ? AND ?")
		args = append(args, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
	}

	if len(req.ServerIds) > 0 {
		conds = append(conds,
			fmt.Sprintf("session.server_id in (%s)", util.IntArrayToString(req.ServerIds, ",")),
		)
	}

	if len(req.MapIds) > 0 {
		conds = append(conds,
			fmt.Sprintf("session.map_id in (%s)", util.IntArrayToString(req.MapIds, ",")),
		)
	}

	if len(req.Statuses) > 0 {
		conds = append(conds,
			fmt.Sprintf("session.status in (%s)", util.IntArrayToString(req.Statuses, ",")),
		)
	}

	if req.Difficulty != nil {
		conds = append(conds, fmt.Sprintf("session.diff = %v", *req.Difficulty))
	}

	if req.Length != nil {
		if *req.Length == models.Custom {
			conds = append(conds, fmt.Sprintf("session.length NOT IN (%v, %v, %v)",
				models.Short, models.Medium, models.Long))
		} else {
			conds = append(conds, fmt.Sprintf("session.length = %v", *req.Length))
		}
	}

	if req.Mode != nil {
		conds = append(conds, fmt.Sprintf("session.mode = %v", *req.Mode))
	}

	if req.Extra != nil {
		filters := req.Extra

		if filters.Wave != nil {
			if stmt, a, ok := filters.Wave.ToStatement("gd.wave"); ok {
				conds = append(conds, stmt)
				args = append(args, a...)
			}
		}

		if filters.Difficulty != nil {
			if stmt, a, ok := filters.Difficulty.ToStatement("diff.final_score"); ok {
				conds = append(conds, stmt)
				args = append(args, a...)
			}
		}

		if filters.MaxMonsters != nil {
			if stmt, a, ok := filters.MaxMonsters.ToStatement("extra.max_monsters"); ok {
				conds = append(conds, stmt)
				args = append(args, a...)
			}
		}

		if filters.SpawnCycle != nil {
			conds = append(conds, "extra.spawn_cycle LIKE ?")
			args = append(args, fmt.Sprintf("%%%v%%", *filters.SpawnCycle))
		}

		if filters.ZedsType != nil {
			conds = append(conds, "extra.zeds_type LIKE ?")
			args = append(args, fmt.Sprintf("%v%%", *filters.ZedsType))
		}
	}

	stmt := fmt.Sprintf(`
		WITH filtered_matches AS (
			SELECT 
				session.id AS session_id,
				coalesce(diff.final_score, 0) AS final_score,
				coalesce(diff.potential_score, 0) AS potential_score,
				session.updated_at AS updated_at
			FROM session
			INNER JOIN session_game_data gd ON gd.session_id = session.id
			LEFT JOIN session_game_data_extra extra ON extra.session_id = session.id
			LEFT JOIN session_diff diff ON diff.session_id = session.id
			WHERE %v
		), pagination AS (
			SELECT session_id
			FROM 
			filtered_matches cte
			ORDER BY %v %v
			LIMIT %v, %v
		), metadata AS (
			SELECT count(*) as count
			FROM filtered_matches
		)
		SELECT
			metadata.count,
			%v
		FROM pagination cte
		INNER JOIN session ON session.id = cte.session_id
		CROSS JOIN metadata`,
		strings.Join(conds, " AND "),
		sortBy, direction,
		page*limit, limit,
		strings.Join(fields, ", "),
	)

	// Execute filter query
	rows, err := s.db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	var total int
	items := []*Match{}

	defer rows.Close()
	for rows.Next() {
		item := Match{}
		sessionData := MatchSession{}

		fields := []any{
			&total,
			&sessionData.Id, &sessionData.ServerId, &sessionData.MapId,
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

		items = append(items, &item)
	}

	items, err = s.handleIncludes(req.Includes, items)
	if err != nil {
		return nil, err
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

func (s *MatchesFilterService) handleIncludes(
	includes *FilterMatchesRequestIncludes,
	items []*Match,
) ([]*Match, error) {
	if includes == nil {
		return items, nil
	}

	sessionIds := s.getSessionIdsFromMatches(items)

	if includes.ServerData != nil && *includes.ServerData {
		serverData, err := s.getServerData(sessionIds)
		if err != nil {
			return nil, err
		}

		for i := range items {
			for j := range serverData {
				if items[i].Session.Id == serverData[j].MatchId {
					items[i].Details.Server = serverData[j].Server
				}
			}
		}
	}

	if includes.MapData != nil && *includes.MapData {
		mapData, err := s.getMapData(sessionIds)
		if err != nil {
			return nil, err
		}

		for i := range items {
			for j := range mapData {
				if items[i].Session.Id == mapData[j].MatchId {
					items[i].Details.Map = mapData[j].Map
				}
			}
		}
	}

	if includes.GameData != nil && *includes.GameData {
		gameData, err := s.getGameData(sessionIds)
		if err != nil {
			return nil, err
		}

		for i := range items {
			for j := range gameData {
				if items[i].Session.Id == gameData[j].MatchId {
					items[i].Details.GameData = gameData[j].GameData
				}
			}
		}
	}

	if includes.ExtraGameData != nil && *includes.ExtraGameData {
		cdData, err := s.getExtraGameData(sessionIds)
		if err != nil {
			return nil, err
		}

		for i := range items {
			for j := range cdData {
				if items[i].Session.Id == cdData[j].MatchId {
					items[i].Details.ExtraGameData = cdData[j].CDData
				}
			}
		}
	}

	if includes.LiveData != nil && *includes.LiveData {
		playerData, err := s.getPlayerData(sessionIds)
		if err != nil {
			return nil, err
		}

		for i := range items {
			items[i].Details.LiveData = &MatchLiveData{}

			for j := range playerData {
				if items[i].Session.Id == playerData[j].MatchId {
					items[i].Details.LiveData.Players = playerData[j].Players
					items[i].Details.LiveData.Spectators = playerData[j].Spectators
				}
			}
		}
	}

	{
		difficulty, err := s.diffService.GetByIds(sessionIds)
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

func (s *MatchesFilterService) getSessionIdsFromMatches(items []*Match) []int {
	sessionIds := []int{}

	for _, item := range items {
		sessionIds = append(sessionIds, item.Session.Id)
	}

	return sessionIds
}
