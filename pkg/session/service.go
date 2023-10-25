package session

import (
	"database/sql"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

type SessionService struct {
	db *sql.DB
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

func (s *SessionService) GetById(id int) (*Session, error) {
	row := s.db.QueryRow(`SELECT * FROM session WHERE id = $1`, id)

	item := Session{}

	err := row.Scan(
		&item.Id, &item.ServerId, &item.MapId,
		&item.Mode, &item.Length, &item.Difficulty,
		&item.Status, &item.CreatedAt, &item.UpdatedAt,
		&item.StartedAt, &item.CompletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *SessionService) GetGameData(id int) (*GameData, error) {
	row := s.db.QueryRow(`
		SELECT
			max_players, players_online, players_alive, 
			wave, is_trader_time, zeds_left 
		FROM session_game_data WHERE session_id = $1`, id,
	)

	item := GameData{}

	err := row.Scan(
		&item.MaxPlayers, &item.PlayersOnline, &item.PlayersAlive,
		&item.Wave, &item.IsTraderTime, &item.ZedsLeft,
	)

	return &item, err
}

func (s *SessionService) GetCDData(id int) (*models.CDGameData, error) {
	row := s.db.QueryRow(`
		SELECT spawn_cycle, max_monsters, wave_size_fakes, zeds_type
		FROM session_game_data_cd WHERE session_id = $1`, id,
	)

	item := models.CDGameData{}

	err := row.Scan(
		&item.SpawnCycle, &item.MaxMonsters,
		&item.WaveSizeFakes, &item.ZedsType,
	)

	return &item, err
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
		UPDATE session
		SET updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`,
		data.SessionId,
	)

	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
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
