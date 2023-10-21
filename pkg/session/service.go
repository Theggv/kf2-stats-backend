package session

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/maps"
	"github.com/theggv/kf2-stats-backend/pkg/server"
)

type SessionService struct {
	db            *sql.DB
	mapsService   *maps.MapsService
	serverService *server.ServerService
}

func (s *SessionService) Inject(
	mapsService *maps.MapsService,
	serverService *server.ServerService,
) {
	s.mapsService = mapsService
	s.serverService = serverService
}

func (s *SessionService) initTables() {
	s.db.Exec(`
	CREATE TABLE IF NOT EXISTS session (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER NOT NULL REFERENCES server(id) 
			ON UPDATE CASCADE 
			ON DELETE CASCADE,
		map_id INTEGER NOT NULL,

		mode INTEGER NOT NULL,
		length INTEGER NOT NULL,
		diff INTEGER NOT NULL,

		status INTEGER NOT NULL DEFAULT 0,

		created_at DATETIME DEFAULT CURRENT_TIMESTAMP, 
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		started_at DATETIME DEFAULT NULL,
		completed_at DATETIME DEFAULT NULL
	);
	
	CREATE TABLE IF NOT EXISTS session_game_data (
		session_id INTEGER PRIMARY KEY REFERENCES session(id)
			ON UPDATE CASCADE 
			ON DELETE CASCADE,

		max_players INTEGER NOT NULL DEFAULT 6,
		players_online INTEGER NOT NULL DEFAULT 0,
		players_alive INTEGER NOT NULL DEFAULT 0,
		
		wave INTEGER NOT NULL DEFAULT 0,
		is_trader_time BOOLEAN NOT NULL DEFAULT 0,
		zeds_left INTEGER NOT NULL DEFAULT 0,

		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS session_game_data_cd (
		session_id INTEGER PRIMARY KEY REFERENCES session(id)
			ON UPDATE CASCADE 
			ON DELETE CASCADE,
		
		spawn_cycle TEXT NOT NULL,
		max_monsters INTEGER NOT NULL,
		wave_size_fakes INTEGER NOT NULL,
		zeds_type TEXT NOT NULL,

		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`)
}

func NewSessionService(db *sql.DB) *SessionService {
	service := SessionService{
		db: db,
	}

	service.initTables()

	return &service
}

func (s *SessionService) Create(req CreateSessionRequest) (int, error) {
	res, err := s.db.Exec(`
		INSERT INTO session (server_id, map_id, mode, length, diff) 
		VALUES ($1, $2, $3, $4, $5)`,
		req.ServerId, req.MapId, req.Mode, req.Length, req.Difficulty,
	)

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = s.db.Exec(`INSERT INTO session_game_data (session_id) VALUES ($1)`, id)

	return int(id), err
}

func (s *SessionService) Filter(req FilterSessionsRequest) (*FilterSessionsResponse, error) {
	page, limit := parsePagination(req.Pager)

	attributes := []string{}
	conditions := []string{}
	joins := []string{}

	// Prepare fields
	attributes = append(attributes,
		"session.id", "session.server_id", "session.map_id",
		"session.mode", "session.length", "session.diff",
		"session.status", "session.created_at", "session.updated_at",
	)

	if req.IncludeMap {
		attributes = append(attributes, "maps.name", "maps.preview")
		joins = append(joins, "LEFT JOIN maps ON maps.id = session.map_id")
	}

	if req.IncludeServer {
		attributes = append(attributes, "server.name", "server.address")
		joins = append(joins, "LEFT JOIN server ON server.id = session.server_id")
	}

	// Prepare filter query
	conditions = append(conditions, "1") // in case if no filters passed

	if len(req.ServerId) > 0 {
		conditions = append(conditions,
			fmt.Sprintf("server_id in (%s)", intArrayToString(req.ServerId, ",")),
		)
	}

	if len(req.MapId) > 0 {
		conditions = append(conditions,
			fmt.Sprintf("map_id in (%s)", intArrayToString(req.MapId, ",")),
		)
	}

	if req.Difficulty != 0 {
		conditions = append(conditions, fmt.Sprintf("diff = %v", req.Difficulty))
	}

	if req.Length != 0 {
		if req.Length == models.Custom {
			conditions = append(conditions, fmt.Sprintf("length NOT IN (%v, %v, %v)",
				models.Short, models.Medium, models.Long))
		} else {
			conditions = append(conditions, fmt.Sprintf("length = %v", req.Length))
		}
	}

	if req.Mode != 0 {
		conditions = append(conditions, fmt.Sprintf("mode = %v", req.Mode))
	}

	sql := fmt.Sprintf(`
		SELECT %v FROM session
		%v
		WHERE %v
		LIMIT %v, %v`,
		strings.Join(attributes, " , "),
		strings.Join(joins, "\n"),
		strings.Join(conditions, " AND "), page*limit, limit,
	)

	// Execute filter query
	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	items := []Session{}

	// Parsing results
	for rows.Next() {
		item := Session{}
		itemMap := SessionMap{}
		itemServer := SessionServer{}

		fields := []any{
			&item.Id, &item.ServerId, &item.MapId,
			&item.Mode, &item.Length, &item.Difficulty,
			&item.Status, &item.CreatedAt, &item.UpdatedAt,
		}

		if req.IncludeMap {
			fields = append(fields, &itemMap.Name, &itemMap.Preview)
		}

		if req.IncludeServer {
			fields = append(fields, &itemServer.Name, &itemServer.Address)
		}

		err := rows.Scan(fields...)
		if err != nil {
			fmt.Print(err)
			continue
		}

		if req.IncludeMap {
			item.Map = &itemMap
		}

		if req.IncludeServer {
			item.Server = &itemServer
		}

		items = append(items, item)
	}

	// Prepare count query
	sql = fmt.Sprintf(`
		SELECT COUNT(*) FROM session
		WHERE %v`,
		strings.Join(conditions, " AND "),
	)

	// Execute count query
	row := s.db.QueryRow(sql)

	// Parsing results
	var total int
	if row.Scan(&total) != nil {
		return nil, err
	}

	return &FilterSessionsResponse{
		Items: items,
		Metadata: models.PaginationResponse{
			Page:           page,
			ResultsPerPage: limit,
			TotalResults:   total,
		},
	}, nil
}

func (s *SessionService) GetById(id int) (*Session, error) {
	row := s.db.QueryRow(`SELECT * FROM session WHERE id = $1`, id)

	item := Session{}

	err := row.Scan(
		&item.Id, &item.ServerId, &item.MapId,
		&item.Mode, &item.Length, &item.Difficulty,
		&item.Status, &item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *SessionService) GetLiveMatches() (*GetLiveMatchesResponse, error) {
	rows, err := s.db.Query(`
		SELECT 
			session.id, session.map_id, session.server_id,
			session.mode, session.length, session.diff, session.started_at,
			gd.max_players, gd.players_online, gd.players_alive, 
			gd.wave, gd.is_trader_time, gd.zeds_left,
			cd.spawn_cycle, cd.max_monsters, cd.wave_size_fakes, cd.zeds_type
		FROM session
		INNER JOIN session_game_data gd ON gd.session_id = session.id
		LEFT JOIN session_game_data_cd cd ON cd.session_id = session.id
		WHERE status = $1`, models.InProgress,
	)

	if err != nil {
		return nil, err
	}

	var (
		mapId    int
		serverId int
		items    []LiveMatch
	)

	for rows.Next() {
		item := LiveMatch{}
		cdData := models.CDGameData{}
		gameData := GameData{}

		err := rows.Scan(
			&item.SessionId, &mapId, &serverId,
			&item.Mode, &item.Length, &item.Difficulty, &item.StartedAt,
			&gameData.MaxPlayers, &gameData.PlayersOnline, &gameData.PlayersAlive,
			&gameData.Wave, &gameData.IsTraderTime, &gameData.ZedsLeft,
			&cdData.SpawnCycle, &cdData.MaxMonsters,
			&cdData.WaveSizeFakes, &cdData.ZedsType,
		)

		if err != nil {
			return nil, err
		}

		mapData, err := s.mapsService.GetById(mapId)
		serverData, err := s.serverService.GetById(serverId)

		if mapData == nil || serverData == nil {
			continue
		}

		item.Map = mapData
		item.Server = serverData
		item.GameData = gameData

		if item.Mode == models.ControlledDifficulty && cdData.SpawnCycle != nil {
			item.CDData = &cdData
		}

		items = append(items, item)
	}

	return &GetLiveMatchesResponse{
		Items: items,
	}, nil
}

func (s *SessionService) GetCurrentServerSession(serverId int) (*LiveMatch, error) {
	row := s.db.QueryRow(`
		SELECT 
			session.id, session.map_id, session.status,
			session.mode, session.length, session.diff, session.started_at,
			gd.max_players, gd.players_online, gd.players_alive, 
			gd.wave, gd.is_trader_time, gd.zeds_left,
			cd.spawn_cycle, cd.max_monsters, cd.wave_size_fakes, cd.zeds_type
		FROM session
		INNER JOIN session_game_data gd ON gd.session_id = session.id
		LEFT JOIN session_game_data_cd cd ON cd.session_id = session.id
		WHERE session.server_id = $1
		ORDER BY session.created_at DESC
		LIMIT 1`,
		serverId,
	)

	var (
		mapId    int
		item     LiveMatch
		cdData   models.CDGameData
		gameData GameData
	)

	err := row.Scan(
		&item.SessionId, &mapId, &item.Status,
		&item.Mode, &item.Length, &item.Difficulty, &item.StartedAt,
		&gameData.MaxPlayers, &gameData.PlayersOnline, &gameData.PlayersAlive,
		&gameData.Wave, &gameData.IsTraderTime, &gameData.ZedsLeft,
		&cdData.SpawnCycle, &cdData.MaxMonsters,
		&cdData.WaveSizeFakes, &cdData.ZedsType,
	)

	if err != nil {
		return nil, err
	}

	mapData, err := s.mapsService.GetById(mapId)

	if mapData == nil {
		return nil, err
	}

	item.Map = mapData
	item.GameData = gameData

	if item.Mode == models.ControlledDifficulty && cdData.SpawnCycle != nil {
		item.CDData = &cdData
	}

	return &item, nil
}

func (s *SessionService) UpdateStatus(data UpdateStatusRequest) error {
	_, err := s.db.Exec(`
		UPDATE session 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2`,
		data.Status, data.Id)

	if err != nil {
		return err
	}

	if data.Status == models.InProgress {
		_, err = s.db.Exec(`
			UPDATE session 
			SET started_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
			WHERE id = $1`, data.Id)
	}

	if data.Status == models.Win ||
		data.Status == models.Lose ||
		data.Status == models.Aborted {
		_, err = s.db.Exec(`
			UPDATE session 
			SET completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
			WHERE id = $1`, data.Id)
	}

	return err
}

func (s *SessionService) UpdateGameData(data UpdateGameDataRequest) error {
	gd := data.GameData

	_, err := s.db.Exec(`
		UPDATE session_game_data
		SET max_players = $1, players_online = $2, players_alive = $3,
			wave = $4, is_trader_time = $5, zeds_left = $6,
			updated_at = CURRENT_TIMESTAMP
		WHERE session_id = $7`,
		gd.MaxPlayers, gd.PlayersOnline, gd.PlayersAlive,
		gd.Wave, gd.IsTraderTime, gd.ZedsLeft,
		data.SessionId,
	)

	if data.CDData == nil {
		return err
	}

	cdData := data.CDData

	_, err = s.db.Exec(`
		INSERT INTO session_game_data_cd 
			(session_id, spawn_cycle, max_monsters, wave_size_fakes, zeds_type)
		VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT(session_id) DO UPDATE SET
			spawn_cycle = $2, max_monsters = $3, wave_size_fakes = $4, zeds_type = $5`,
		data.SessionId, cdData.SpawnCycle, cdData.MaxMonsters, cdData.WaveSizeFakes, cdData.ZedsType,
	)

	return err
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
