package matches

import (
	"database/sql"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/steamapi"
	"github.com/theggv/kf2-stats-backend/pkg/maps"
	"github.com/theggv/kf2-stats-backend/pkg/server"
	"github.com/theggv/kf2-stats-backend/pkg/session"
	"github.com/theggv/kf2-stats-backend/pkg/stats"
	"github.com/theggv/kf2-stats-backend/pkg/users"
)

type getMatchWavesPlayersResponse struct {
	WaveId int
	Player *MatchWavePlayer
}

type getMatchWavesPlayersStatsResponse struct {
	PlayerStatsId int
	Stats         *MatchWavePlayerStats
}

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

func (s *MatchesService) GetById(id int) (*Match, error) {
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

func (s *MatchesService) GetLastServerMatch(id int) (*Match, error) {
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

	return s.GetById(sessionId)
}

func (s *MatchesService) GetMatchWaves(sessionId int) (*GetMatchWavesResponse, error) {
	waves, err := s.getMatchWavesOverview(sessionId)
	if err != nil {
		return nil, err
	}

	players, err := s.getMatchWavesPlayers(sessionId)
	if err != nil {
		return nil, err
	}

	playerStats, err := s.getMatchWavesPlayersStats(sessionId)
	if err != nil {
		return nil, err
	}

	// Join stats with players
	for i := 0; i < len(players); i++ {
		for j := 0; j < len(playerStats); j++ {
			if players[i].Player.PlayerStatsId == playerStats[j].PlayerStatsId {
				players[i].Player.Stats = playerStats[j].Stats
			}
		}
	}

	// Join players with waves
	for i := 0; i < len(waves); i++ {
		for j := 0; j < len(players); j++ {
			if waves[i].WaveId == players[j].WaveId {
				waves[i].Players = append(waves[i].Players, players[j].Player)
			}
		}
	}

	// Get unique user ids
	userId := []int{}
	userIdSet := make(map[int]bool)

	for _, player := range players {
		userIdSet[player.Player.UserId] = true
	}

	for key := range userIdSet {
		userId = append(userId, key)
	}

	users, err := s.getUserProfiles(userId)
	if err != nil {
		return nil, err
	}

	return &GetMatchWavesResponse{
		Waves: waves,
		Users: users,
	}, nil
}

func (s *MatchesService) getMatchWavesOverview(sessionId int) ([]*MatchWave, error) {
	rows, err := s.db.Query(`
		SELECT 
			ws.id, ws.wave, ws.attempt, ws.started_at, ws.completed_at
		FROM session
		INNER JOIN wave_stats ws on ws.session_id = session.id
		WHERE session.id = ?
		GROUP BY ws.id 
		ORDER BY ws.id`,
		sessionId,
	)

	if err != nil {
		return nil, err
	}

	// Get waves data
	waves := []*MatchWave{}
	for rows.Next() {
		wave := MatchWave{}

		err := rows.Scan(&wave.WaveId, &wave.Wave, &wave.Attempt, &wave.StartedAt, &wave.CompletedAt)
		if err != nil {
			return nil, err
		}

		waves = append(waves, &wave)
	}

	return waves, nil
}

func (s *MatchesService) getMatchWavesPlayers(sessionId int) ([]*getMatchWavesPlayersResponse, error) {
	rows, err := s.db.Query(`
		SELECT 
			wsp.player_id, ws.id, wsp.id,
			wsp.perk, wsp.level, wsp.prestige, wsp.is_dead
		FROM session
		INNER JOIN wave_stats ws on ws.session_id = session.id
		INNER JOIN wave_stats_player wsp on wsp.stats_id = ws.id
		WHERE session.id = ?
		GROUP BY wsp.id`,
		sessionId,
	)

	if err != nil {
		return nil, err
	}

	result := []*getMatchWavesPlayersResponse{}

	for rows.Next() {
		p := getMatchWavesPlayersResponse{
			Player: &MatchWavePlayer{},
		}

		err := rows.Scan(
			&p.Player.UserId, &p.WaveId, &p.Player.PlayerStatsId,
			&p.Player.Perk, &p.Player.Level, &p.Player.Prestige, &p.Player.IsDead,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, &p)
	}

	return result, nil
}

func (s *MatchesService) getMatchWavesPlayersStats(sessionId int) (
	[]*getMatchWavesPlayersStatsResponse, error,
) {
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
		FROM session
		INNER JOIN wave_stats ws on ws.session_id = session.id
		INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
		INNER JOIN wave_stats_player_kills kills ON kills.player_stats_id = wsp.id
		WHERE session.id = ?`, sessionId,
	)

	if err != nil {
		return nil, err
	}

	players := []*getMatchWavesPlayersStatsResponse{}

	for rows.Next() {
		var (
			unused        int
			playerStatsId int
		)
		s := MatchWavePlayerStats{}
		kills := stats.ZedCounter{}

		err := rows.Scan(&playerStatsId,
			&s.ShotsFired, &s.ShotsHit, &s.ShotsHS,
			&s.DoshEarned, &s.HealsGiven, &s.HealsReceived,
			&s.DamageDealt, &s.DamageTaken,
			&s.ZedTimeCount, &s.ZedTimeLength,
			&unused,
			&kills.Cyst, &kills.AlphaClot, &kills.Slasher, &kills.Stalker, &kills.Crawler, &kills.Gorefast,
			&kills.Rioter, &kills.EliteCrawler, &kills.Gorefiend,
			&kills.Siren, &kills.Bloat, &kills.Edar,
			&kills.Husk, &s.HuskBackpackKills, &s.HuskRages,
			&kills.Scrake, &kills.FP, &kills.QP, &kills.Boss, &kills.Custom,
		)

		if err != nil {
			return nil, err
		}

		s.Kills = kills

		players = append(players, &getMatchWavesPlayersStatsResponse{
			PlayerStatsId: playerStatsId,
			Stats:         &s,
		})
	}

	return players, nil
}

func (s *MatchesService) getUserProfiles(userId []int) (
	[]*models.UserProfile, error,
) {
	users, err := s.userService.GetManyById(userId)
	if err != nil {
		return nil, err
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

func (s *MatchesService) GetMatchPlayerStats(sessionId, userId int) (*GetMatchPlayerStatsResponse, error) {
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

func (s *MatchesService) GetMatchAggregatedStats(sessionId int) (*GetMatchAggregatedStatsResponse, error) {
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

// Deprecated, needs rework
func (s *MatchesService) GetMatchLiveData(sessionId int) (*GetMatchLiveDataResponse, error) {
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
