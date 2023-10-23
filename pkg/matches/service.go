package matches

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/steamapi"
	"github.com/theggv/kf2-stats-backend/pkg/maps"
	"github.com/theggv/kf2-stats-backend/pkg/server"
	"github.com/theggv/kf2-stats-backend/pkg/session"
	"github.com/theggv/kf2-stats-backend/pkg/stats"
	"github.com/theggv/kf2-stats-backend/pkg/users"
)

type MatchesService struct {
	db *sql.DB

	sessionService  *session.SessionService
	mapsService     *maps.MapsService
	serverService   *server.ServerService
	steamApiService *steamapi.SteamApiUserService
}

func (s *MatchesService) Inject(
	sessionService *session.SessionService,
	mapsService *maps.MapsService,
	serverService *server.ServerService,
	steamApiService *steamapi.SteamApiUserService,
) {
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
	if err == nil {
		match.CDData = cdData
	}

	return &match, nil
}

func (s *MatchesService) getLastServerMatch(id int) (*Match, error) {
	row := s.db.QueryRow(`
		SELECT server_id FROM session
		WHERE server_id = $1
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
	page, limit := parsePagination(req.Pager)

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
		joins = append(joins, "LEFT JOIN maps ON maps.id = session.map_id")
	}

	if req.IncludeServer {
		attributes = append(attributes, "server.name", "server.address")
		joins = append(joins, "LEFT JOIN server ON server.id = session.server_id")
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
			fmt.Sprintf("session.server_id in (%s)", intArrayToString(req.ServerId, ",")),
		)
	}

	if len(req.MapId) > 0 {
		conditions = append(conditions,
			fmt.Sprintf("session.map_id in (%s)", intArrayToString(req.MapId, ",")),
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
		gameData := session.GameData{}
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

	// Prepare count query
	sql = fmt.Sprintf(`
		SELECT count(*) FROM session
		%v
		WHERE %v`,
		strings.Join(joins, "\n"),
		strings.Join(conditions, " AND "),
	)

	// Execute count query
	row := s.db.QueryRow(sql)

	// Parsing results
	var total int
	if row.Scan(&total) != nil {
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

func parsePagination(pager models.PaginationRequest) (int, int) {
	page := pager.Page
	resultsPerPage := pager.ResultsPerPage

	if page < 0 {
		page = 0
	}

	if resultsPerPage < 10 {
		resultsPerPage = 10
	}

	if resultsPerPage > 100 {
		resultsPerPage = 100
	}

	return page, resultsPerPage
}

func intArrayToString(a []int, delimiter string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delimiter, -1), "[]")
}

func (s *MatchesService) getMatchStats(sessionId int) (*GetMatchStatsResponse, error) {
	rows, err := s.db.Query(`
		SELECT 
			view_indexes.wave_stats_id,
			ws.wave, ws.attempt, ws.started_at, ws.completed_at
		FROM view_indexes
		INNER JOIN wave_stats ws on ws.id = view_indexes.wave_stats_id
		WHERE view_indexes.session_id = $1
		GROUP BY wave_stats_id 
		ORDER BY wave_stats_id`, sessionId,
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

	return &GetMatchStatsResponse{
		Waves: waves,
	}, nil
}

func (s *MatchesService) getWavePlayers(waveId int) ([]Player, error) {
	rows, err := s.db.Query(`
		SELECT 
			users.id, users.auth_id, users.auth_type, users.name,
			view_indexes.wave_stats_player_id,
			wsp.perk, wsp.level, wsp.prestige, wsp.is_dead
		FROM view_indexes
		INNER JOIN wave_stats_player wsp on wsp.id = view_indexes.wave_stats_player_id
		INNER JOIN users on users.id = wsp.player_id
		WHERE view_indexes.wave_stats_id = $1
		GROUP BY view_indexes.wave_stats_player_id
		ORDER BY users.id`,
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
			if player.AuthType != users.Steam {
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
			wave_stats_player.id,
			wave_stats_player.shots_fired,
			wave_stats_player.shots_hit,
			wave_stats_player.shots_hs,
			wave_stats_player.dosh_earned,
			wave_stats_player.heals_given,
			wave_stats_player.heals_recv,
			wave_stats_player.damage_dealt,
			wave_stats_player.damage_taken,
			wave_stats_player.zedtime_count,
			wave_stats_player.zedtime_length,
			kills.*,
			injured_by.*
		FROM view_indexes
		INNER JOIN wave_stats_player ON wave_stats_player.id = view_indexes.wave_stats_player_id
		INNER JOIN wave_stats_player_kills kills ON kills.player_stats_id = wave_stats_player.id
		INNER JOIN wave_stats_player_injured_by injured_by ON injured_by.player_stats_id = wave_stats_player.id
		WHERE view_indexes.wave_stats_id = $1`, waveId,
	)

	if err != nil {
		return nil, err
	}

	players := []PlayerWaveStats{}

	for rows.Next() {
		var useless int
		player := PlayerWaveStats{}
		kills := stats.ZedCounter{}
		injuredBy := stats.ZedCounter{}

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
			&kills.Scrake, &kills.FP, &kills.QP, &kills.Boss,
			&useless,
			&injuredBy.Cyst, &injuredBy.AlphaClot, &injuredBy.Slasher, &injuredBy.Stalker, &injuredBy.Crawler, &injuredBy.Gorefast,
			&injuredBy.Rioter, &injuredBy.EliteCrawler, &injuredBy.Gorefiend,
			&injuredBy.Siren, &injuredBy.Bloat, &injuredBy.Edar, &injuredBy.Husk,
			&injuredBy.Scrake, &injuredBy.FP, &injuredBy.QP, &injuredBy.Boss,
		)

		if err != nil {
			return nil, err
		}

		player.Kills = kills
		player.Injuredby = injuredBy

		players = append(players, player)
	}

	return &GetMatchWaveStatsResponse{
		Players: players,
	}, nil
}

func (s *MatchesService) getMatchPlayerStats(sessionId, userId int) (*GetMatchPlayerStatsResponse, error) {
	rows, err := s.db.Query(`
		SELECT
			wave_stats_player.id,
			wave_stats_player.shots_fired,
			wave_stats_player.shots_hit,
			wave_stats_player.shots_hs,
			wave_stats_player.dosh_earned,
			wave_stats_player.heals_given,
			wave_stats_player.heals_recv,
			wave_stats_player.damage_dealt,
			wave_stats_player.damage_taken,
			wave_stats_player.zedtime_count,
			wave_stats_player.zedtime_length,
			kills.*,
			injured_by.*
		FROM view_indexes
		INNER JOIN wave_stats_player ON wave_stats_player.id = view_indexes.wave_stats_player_id
		INNER JOIN wave_stats_player_kills kills ON kills.player_stats_id = wave_stats_player.id
		INNER JOIN wave_stats_player_injured_by injured_by ON injured_by.player_stats_id = wave_stats_player.id
		WHERE view_indexes.session_id = $1 and wave_stats_player.player_id = $2`, sessionId, userId,
	)

	if err != nil {
		return nil, err
	}

	waves := []PlayerWaveStats{}

	for rows.Next() {
		var useless int
		player := PlayerWaveStats{}
		kills := stats.ZedCounter{}
		injuredBy := stats.ZedCounter{}

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
			&kills.Scrake, &kills.FP, &kills.QP, &kills.Boss,
			&useless,
			&injuredBy.Cyst, &injuredBy.AlphaClot, &injuredBy.Slasher, &injuredBy.Stalker, &injuredBy.Crawler, &injuredBy.Gorefast,
			&injuredBy.Rioter, &injuredBy.EliteCrawler, &injuredBy.Gorefiend,
			&injuredBy.Siren, &injuredBy.Bloat, &injuredBy.Edar, &injuredBy.Husk,
			&injuredBy.Scrake, &injuredBy.FP, &injuredBy.QP, &injuredBy.Boss,
		)

		if err != nil {
			return nil, err
		}

		player.Kills = kills
		player.Injuredby = injuredBy

		waves = append(waves, player)
	}

	return &GetMatchPlayerStatsResponse{
		Waves: waves,
	}, nil
}
