package matches

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
	"github.com/theggv/kf2-stats-backend/pkg/stats"
	"github.com/theggv/kf2-stats-backend/pkg/users"
)

type MatchesService struct {
	db *sql.DB

	userService     *users.UserService
	sessionService  *session.SessionService
	mapsService     *maps.MapsService
	serverService   *server.ServerService
	steamApiService *steamapi.SteamApiUserService
}

func (s *MatchesService) Inject(
	userService *users.UserService,
	sessionService *session.SessionService,
	mapsService *maps.MapsService,
	serverService *server.ServerService,
	steamApiService *steamapi.SteamApiUserService,
) {
	s.userService = userService
	s.sessionService = sessionService
	s.mapsService = mapsService
	s.serverService = serverService
	s.steamApiService = steamApiService
}

func NewMatchesService(db *sql.DB) *MatchesService {
	service := MatchesService{
		db: db,
	}

	service.setupTasks()

	return &service
}

func (s *MatchesService) getById(id int) (*Match, error) {
	match := Match{}

	session, err := s.sessionService.GetById(id)
	if err != nil {
		return nil, err
	}
	match.Session = MatchSession{
		SessionId:   session.Id,
		ServerId:    session.ServerId,
		MapId:       session.MapId,
		Mode:        session.Mode,
		Length:      session.Length,
		Difficulty:  session.Difficulty,
		Status:      session.Status,
		CreatedAt:   session.CreatedAt,
		UpdatedAt:   session.UpdatedAt,
		StartedAt:   session.StartedAt,
		CompletedAt: session.CompletedAt,
	}

	mapData, err := s.mapsService.GetById(session.MapId)
	if err == nil {
		match.Map = &MatchMap{
			Name:    &mapData.Name,
			Preview: &mapData.Preview,
		}
	}

	serverData, err := s.serverService.GetById(session.ServerId)
	if err == nil {
		match.Server = &MatchServer{
			Name:    &serverData.Name,
			Address: &serverData.Address,
		}
	}

	gameData, err := s.sessionService.GetGameData(session.Id)
	if err == nil {
		match.GameData = gameData
	}

	cdData, err := s.sessionService.GetCDData(session.Id)
	if err == nil && cdData.SpawnCycle != nil {
		match.CDData = cdData
	}

	return &match, nil
}

func (s *MatchesService) getLastServerMatch(id int) (*Match, error) {
	row := s.db.QueryRow(`
		SELECT session.id FROM session
		WHERE server_id = ?
		ORDER BY id desc
		LIMIT 1`, id,
	)

	var sessionId int
	err := row.Scan(&sessionId)

	if err != nil {
		return nil, err
	}

	return s.getById(sessionId)
}

func (s *MatchesService) filter(req FilterMatchesRequest) (*FilterMatchesResponse, error) {
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

	if req.IncludeGameData {
		attributes = append(attributes,
			"gd.max_players", "gd.players_online", "gd.players_alive",
			"gd.wave", "gd.is_trader_time", "gd.zeds_left",
		)
		joins = append(joins, "INNER JOIN session_game_data gd ON session.id = gd.session_id")
	}

	if req.IncludeMap {
		attributes = append(attributes, "maps.name", "maps.preview")
		joins = append(joins, "INNER JOIN maps ON maps.id = session.map_id")
	}

	if req.IncludeServer {
		attributes = append(attributes, "server.name", "server.address")
		joins = append(joins, "INNER JOIN server ON server.id = session.server_id")
	}

	if req.IncludeCDData {
		attributes = append(attributes,
			"cd.spawn_cycle", "cd.max_monsters",
			"cd.wave_size_fakes", "cd.zeds_type",
		)
		joins = append(joins, "LEFT JOIN session_game_data_cd cd ON session.id = cd.session_id")
	}

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

	if req.Status != nil {
		conditions = append(conditions, fmt.Sprintf("session.status = %v", *req.Status))
	}

	// Order
	if req.ReverseOrder != nil && *req.ReverseOrder == true {
		order = "desc"
	}

	sql := fmt.Sprintf(`
		SELECT %v FROM session
		%v
		WHERE %v
		ORDER BY session.id %v
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

	defer rows.Close()
	items := []Match{}

	// Parsing results
	for rows.Next() {
		item := Match{}
		sessionData := MatchSession{}
		mapData := MatchMap{}
		serverData := MatchServer{}
		gameData := models.GameData{}
		cdData := models.CDGameData{}

		fields := []any{
			&sessionData.SessionId, &sessionData.ServerId, &sessionData.MapId,
			&sessionData.Mode, &sessionData.Length, &sessionData.Difficulty,
			&sessionData.Status, &sessionData.CreatedAt, &sessionData.UpdatedAt,
			&sessionData.StartedAt, &sessionData.CompletedAt,
		}

		if req.IncludeGameData {
			fields = append(fields,
				&gameData.MaxPlayers, &gameData.PlayersOnline, &gameData.PlayersAlive,
				&gameData.Wave, &gameData.IsTraderTime, &gameData.ZedsLeft,
			)
		}

		if req.IncludeMap {
			fields = append(fields, &mapData.Name, &mapData.Preview)
		}

		if req.IncludeServer {
			fields = append(fields, &serverData.Name, &serverData.Address)
		}

		if req.IncludeCDData {
			fields = append(fields,
				&cdData.SpawnCycle, &cdData.MaxMonsters,
				&cdData.WaveSizeFakes, &cdData.ZedsType,
			)
		}

		err := rows.Scan(fields...)
		if err != nil {
			fmt.Print(err)
			continue
		}

		item.Session = sessionData

		if req.IncludeGameData {
			item.GameData = &gameData
		}

		if req.IncludeMap {
			item.Map = &mapData
		}

		if req.IncludeServer {
			item.Server = &serverData
		}

		if req.IncludeCDData && cdData.SpawnCycle != nil {
			item.CDData = &cdData
		}

		items = append(items, item)
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

func (s *MatchesService) getMatchWaves(sessionId int) (*GetMatchWavesResponse, error) {
	rows, err := s.db.Query(`
		SELECT 
			ws.id, ws.wave, ws.attempt, ws.started_at, ws.completed_at
		FROM session
		INNER JOIN wave_stats ws on ws.session_id = session.id
		WHERE session.id = ?
		GROUP BY ws.id 
		ORDER BY ws.id`, sessionId,
	)

	if err != nil {
		return nil, err
	}

	waves := []MatchWave{}
	players := [][]Player{}

	for rows.Next() {
		wave := MatchWave{}

		err := rows.Scan(&wave.Id, &wave.Wave, &wave.Attempt, &wave.StartedAt, &wave.CompletedAt)
		if err != nil {
			return nil, err
		}

		wavePlayers, err := s.getWavePlayers(wave.Id)
		if err != nil {
			return nil, err
		}

		players = append(players, wavePlayers)
		waves = append(waves, wave)
	}

	playersWithExtraData, err := s.getPlayersSteamData(&players)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(waves); i++ {
		waves[i].Players = (*playersWithExtraData)[i]
	}

	return &GetMatchWavesResponse{
		Waves: waves,
	}, nil
}

func (s *MatchesService) getWavePlayers(waveId int) ([]Player, error) {
	rows, err := s.db.Query(`
		SELECT 
			users.id, users.auth_id, users.auth_type, users.name,
			wsp.id,
			wsp.perk, wsp.level, wsp.prestige, wsp.is_dead
		FROM wave_stats ws
		INNER JOIN wave_stats_player wsp on wsp.stats_id = ws.id
		INNER JOIN users on users.id = wsp.player_id
		WHERE ws.id = ?
		GROUP BY wsp.id`,
		waveId,
	)

	if err != nil {
		return nil, err
	}

	players := []Player{}

	for rows.Next() {
		p := Player{}

		err := rows.Scan(
			&p.Id, &p.AuthId, &p.AuthType, &p.Name,
			&p.PlayerStatsId, &p.Perk, &p.Level, &p.Prestige, &p.IsDead,
		)
		if err != nil {
			return nil, err
		}

		players = append(players, p)
	}

	return players, nil
}

func (s *MatchesService) getPlayersSteamData(players *[][]Player) (*[][]PlayerWithSteamData, error) {
	set := make(map[string]bool)
	steamIds := []string{}

	for _, wavePlayers := range *players {
		for _, player := range wavePlayers {
			if player.AuthType != models.Steam {
				continue
			}
			set[player.AuthId] = true
		}
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

	playersWithSteamData := [][]PlayerWithSteamData{}

	for _, wavePlayers := range *players {
		wavePlayersWithSteamData := []PlayerWithSteamData{}

		for _, player := range wavePlayers {
			playerWithSteamData := PlayerWithSteamData{
				Id:            player.Id,
				Name:          player.Name,
				PlayerStatsId: player.PlayerStatsId,
				Perk:          player.Perk,
				Level:         player.Level,
				Prestige:      player.Prestige,
				IsDead:        player.IsDead,
			}

			if data, exists := steamDataSet[player.AuthId]; exists {
				playerWithSteamData.ProfileUrl = &data.ProfileUrl
				playerWithSteamData.Avatar = &data.Avatar
			}

			wavePlayersWithSteamData = append(wavePlayersWithSteamData, playerWithSteamData)
		}

		playersWithSteamData = append(playersWithSteamData, wavePlayersWithSteamData)
	}

	return &playersWithSteamData, nil
}

func (s *MatchesService) getWavePlayersStats(waveId int) (*GetMatchWaveStatsResponse, error) {
	rows, err := s.db.Query(`
		SELECT
			wsp.id,
			wsp.shots_fired,
			wsp.shots_hit,
			wsp.shots_hs,
			wsp.dosh_earned,
			wsp.heals_given,
			wsp.heals_recv,
			wsp.damage_dealt,
			wsp.damage_taken,
			wsp.zedtime_count,
			wsp.zedtime_length,
			kills.*
		FROM wave_stats ws
		INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
		INNER JOIN wave_stats_player_kills kills ON kills.player_stats_id = wsp.id
		WHERE ws.id = ?`, waveId,
	)

	if err != nil {
		return nil, err
	}

	players := []PlayerWaveStats{}

	for rows.Next() {
		var useless int
		player := PlayerWaveStats{}
		kills := stats.ZedCounter{}

		err := rows.Scan(&player.PlayerStatsId,
			&player.ShotsFired, &player.ShotsHit, &player.ShotsHS,
			&player.DoshEarned, &player.HealsGiven, &player.HealsReceived,
			&player.DamageDealt, &player.DamageTaken,
			&player.ZedTimeCount, &player.ZedTimeLength,
			&useless,
			&kills.Cyst, &kills.AlphaClot, &kills.Slasher, &kills.Stalker, &kills.Crawler, &kills.Gorefast,
			&kills.Rioter, &kills.EliteCrawler, &kills.Gorefiend,
			&kills.Siren, &kills.Bloat, &kills.Edar,
			&kills.Husk, &player.HuskBackpackKills, &player.HuskRages,
			&kills.Scrake, &kills.FP, &kills.QP, &kills.Boss, &kills.Custom,
		)

		if err != nil {
			return nil, err
		}

		player.Kills = kills

		players = append(players, player)
	}

	return &GetMatchWaveStatsResponse{
		Players: players,
	}, nil
}

func (s *MatchesService) getMatchPlayerStats(sessionId, userId int) (*GetMatchPlayerStatsResponse, error) {
	rows, err := s.db.Query(`
		SELECT
			wsp.id,
			wsp.shots_fired,
			wsp.shots_hit,
			wsp.shots_hs,
			wsp.dosh_earned,
			wsp.heals_given,
			wsp.heals_recv,
			wsp.damage_dealt,
			wsp.damage_taken,
			wsp.zedtime_count,
			wsp.zedtime_length,
			kills.*
		FROM wave_stats ws
		INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
		INNER JOIN wave_stats_player_kills kills ON kills.player_stats_id = wsp.id
		WHERE ws.session_id = ? and wsp.player_id = ?`, sessionId, userId,
	)

	if err != nil {
		return nil, err
	}

	waves := []PlayerWaveStats{}

	for rows.Next() {
		var useless int
		player := PlayerWaveStats{}
		kills := stats.ZedCounter{}

		err := rows.Scan(&player.PlayerStatsId,
			&player.ShotsFired, &player.ShotsHit, &player.ShotsHS,
			&player.DoshEarned, &player.HealsGiven, &player.HealsReceived,
			&player.DamageDealt, &player.DamageTaken,
			&player.ZedTimeCount, &player.ZedTimeLength,
			&useless,
			&kills.Cyst, &kills.AlphaClot, &kills.Slasher, &kills.Stalker, &kills.Crawler, &kills.Gorefast,
			&kills.Rioter, &kills.EliteCrawler, &kills.Gorefiend,
			&kills.Siren, &kills.Bloat, &kills.Edar,
			&kills.Husk, &player.HuskBackpackKills, &player.HuskRages,
			&kills.Scrake, &kills.FP, &kills.QP, &kills.Boss, &kills.Custom,
		)

		if err != nil {
			return nil, err
		}

		player.Kills = kills

		waves = append(waves, player)
	}

	return &GetMatchPlayerStatsResponse{
		Waves: waves,
	}, nil
}

func (s *MatchesService) getMatchAggregatedStats(sessionId int) (*GetMatchAggregatedStatsResponse, error) {
	rows, err := s.db.Query(`
		SELECT
			wsp.player_id,
			sum(timestampdiff(MINUTE, ws.started_at, ws.completed_at)) as playtime,
			sum(shots_fired), sum(shots_hit), sum(shots_hs),
			sum(dosh_earned), sum(heals_given), sum(heals_recv),
			sum(damage_dealt), sum(damage_taken),
			sum(zedtime_count), sum(zedtime_length),
			sum(aggr_kills.total), sum(aggr_kills.large), sum(husk_r)
		FROM wave_stats ws
		INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
		INNER JOIN wave_stats_player_kills kills ON kills.player_stats_id = wsp.id
		INNER JOIN aggregated_kills aggr_kills ON aggr_kills.player_stats_id = wsp.id
		WHERE ws.session_id = ?
		GROUP BY wsp.player_id`, sessionId,
	)

	if err != nil {
		return nil, err
	}

	players := []AggregatedPlayerStats{}
	for rows.Next() {
		stats := AggregatedPlayerStats{}

		err = rows.Scan(&stats.UserId, &stats.PlayTime,
			&stats.ShotsFired, &stats.ShotsHit, &stats.ShotsHS,
			&stats.DoshEarned, &stats.HealsGiven, &stats.HealsReceived,
			&stats.DamageDealt, &stats.DamageTaken,
			&stats.ZedTimeCount, &stats.ZedTimeLength,
			&stats.Kills, &stats.LargeKills, &stats.HuskRages,
		)

		players = append(players, stats)
	}

	return &GetMatchAggregatedStatsResponse{
		Players: players,
	}, nil
}

func (s *MatchesService) getMatchLiveData(sessionId int) (*GetMatchLiveDataResponse, error) {
	sql := `
		SELECT
			max_players, players_online, players_alive, 
			wave, is_trader_time, zeds_left,
			spawn_cycle, max_monsters, wave_size_fakes, zeds_type
		FROM session_game_data gd
		LEFT JOIN session_game_data_cd cd ON cd.session_id = gd.session_id
		WHERE gd.session_id = ?`

	gameData := models.GameData{}
	cdData := models.CDGameData{}

	err := s.db.QueryRow(sql, sessionId).Scan(
		&gameData.MaxPlayers, &gameData.PlayersOnline, &gameData.PlayersAlive,
		&gameData.Wave, &gameData.IsTraderTime, &gameData.ZedsLeft,
		&cdData.SpawnCycle, &cdData.MaxMonsters,
		&cdData.WaveSizeFakes, &cdData.ZedsType,
	)
	if err != nil {
		return nil, err
	}

	if gameData.IsTraderTime {
		gameData.ZedsLeft = 0
	}

	var status models.GameStatus
	err = s.db.QueryRow(`SELECT status FROM session WHERE id = ?`, sessionId).Scan(&status)

	res := GetMatchLiveDataResponse{
		Status:     status,
		GameData:   gameData,
		Players:    []*GetMatchLiveDataResponsePlayer{},
		Spectators: []*GetMatchLiveDataResponsePlayer{},
	}
	if cdData.MaxMonsters != nil {
		res.CDData = &cdData
	}

	rows, err := s.db.Query(`
		SELECT 
			users.id,
			users.name,
			users.auth_id,
			users.auth_type,
			activity.perk,
			activity.level,
			activity.prestige,
			activity.health,
			activity.armor,
			activity.is_spectator
		FROM users
		INNER JOIN users_activity activity ON users.id = activity.user_id
		where current_session_id = ?
		`, sessionId,
	)

	steamIdSet := make(map[string]bool)
	for rows.Next() {
		item := GetMatchLiveDataResponsePlayer{}

		err = rows.Scan(
			&item.Id, &item.Name, &item.AuthId, &item.AuthType,
			&item.Perk, &item.Level, &item.Prestige,
			&item.Health, &item.Armor,
			&item.IsSpectator,
		)

		if item.AuthType == models.Steam {
			steamIdSet[item.AuthId] = true
		}

		if item.IsSpectator {
			res.Spectators = append(res.Spectators, &item)
		} else {
			res.Players = append(res.Players, &item)
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

		for _, item := range res.Players {
			if data, ok := steamData[item.AuthId]; ok {
				item.Avatar = &data.Avatar
				item.ProfileUrl = &data.ProfileUrl
			}
		}

		for _, item := range res.Spectators {
			if data, ok := steamData[item.AuthId]; ok {
				item.Avatar = &data.Avatar
				item.ProfileUrl = &data.ProfileUrl
			}
		}
	}

	return &res, nil
}
