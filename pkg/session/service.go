package session

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"

	"github.com/theggv/kf2-stats-backend/pkg/common/demorecord"
	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
	"github.com/theggv/kf2-stats-backend/pkg/maps"
	"github.com/theggv/kf2-stats-backend/pkg/server"
	"github.com/theggv/kf2-stats-backend/pkg/users"
)

type SessionService struct {
	db *sql.DB

	mapsService   *maps.MapsService
	serverService *server.ServerService
	usersService  *users.UserService
}

func NewSessionService(db *sql.DB) *SessionService {
	service := SessionService{
		db: db,
	}

	return &service
}

func (s *SessionService) Inject(
	mapsService *maps.MapsService,
	serverService *server.ServerService,
	usersService *users.UserService,
) {
	s.mapsService = mapsService
	s.serverService = serverService
	s.usersService = usersService
}

func (s *SessionService) Create(req CreateSessionRequest) (int, error) {
	serverId, err := s.serverService.Create(server.AddServerRequest{
		Name:    req.ServerName,
		Address: req.ServerAddress,
	})

	if err != nil {
		return 0, err
	}

	mapId, err := s.mapsService.Create(maps.AddMapRequest{
		Name: req.MapName,
	})

	if err != nil {
		return 0, err
	}

	res, err := s.db.Exec(`
		INSERT INTO session (server_id, map_id, mode, length, diff) 
		VALUES (?, ?, ?, ?, ?)`,
		serverId, mapId, req.Mode, req.Length, req.Difficulty,
	)

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = s.db.Exec(`INSERT INTO session_game_data (session_id) VALUES (?)`, id)

	return int(id), err
}

func (s *SessionService) GetById(id int) (*Session, error) {
	row := s.db.QueryRow(`
		SELECT 
			id, server_id, map_id,
			mode, length, diff, status,
			created_at, updated_at, started_at, completed_at
		FROM session WHERE id = ?`, id,
	)

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

func (s *SessionService) GetGameData(id int) (*models.GameData, error) {
	row := s.db.QueryRow(`
		SELECT
			max_players, players_online, players_alive, 
			wave, is_trader_time, zeds_left 
		FROM session_game_data WHERE session_id = ?`, id,
	)

	item := models.GameData{}

	err := row.Scan(
		&item.MaxPlayers, &item.PlayersOnline, &item.PlayersAlive,
		&item.Wave, &item.IsTraderTime, &item.ZedsLeft,
	)

	return &item, err
}

func (s *SessionService) GetCDData(id int) (*models.ExtraGameData, error) {
	row := s.db.QueryRow(`
		SELECT spawn_cycle, max_monsters, wave_size_fakes, zeds_type
		FROM session
		INNER JOIN wave_stats ws on ws.session_id = session.id
		INNER JOIN wave_stats_extra cd on cd.stats_id = ws.id
		WHERE session.id = ? and ws.wave <= session.length
		ORDER BY ws.id DESC
		LIMIT 1`, id,
	)

	item := models.ExtraGameData{}

	err := row.Scan(
		&item.SpawnCycle, &item.MaxMonsters,
		&item.WaveSizeFakes, &item.ZedsType,
	)

	return &item, err
}

func (s *SessionService) UpdateStatus(data UpdateStatusRequest) error {
	_, err := s.db.Exec(`
		UPDATE session 
		SET status = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`,
		data.Status, data.Id)

	if err != nil {
		return err
	}

	if data.Status == models.InProgress {
		_, err = s.db.Exec(`
			UPDATE session 
			SET started_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
			WHERE id = ?`, data.Id)
	}

	if data.Status == models.Win ||
		data.Status == models.Lose ||
		data.Status == models.Aborted {
		s.db.Exec(`
			UPDATE session 
			SET completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
			WHERE id = ?`, data.Id)

		row := s.db.QueryRow(`
			SELECT count(*)
			FROM session
			INNER JOIN wave_stats ws ON ws.session_id = session.id
			INNER JOIN wave_stats_player wsp ON wsp.stats_id = ws.id
			WHERE session.id = ?`, data.Id)

		var count int
		err := row.Scan(&count)
		if err != nil {
			return err
		}

		if count == 0 {
			s.db.Exec(`DELETE FROM session WHERE id = ?`, data.Id)
		}
	}

	return err
}

func (s *SessionService) UploadDemo(raw []byte) error {
	demo, err := demorecord.Parse(raw)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	writer := gzip.NewWriter(&b)
	writer.Write(raw)
	writer.Close()

	_, err = s.db.Exec(`
		INSERT INTO session_demo (session_id, data) VALUES 
		(?, ?)`,
		demo.Header.SessionId, b.Bytes(),
	)

	return err
}

func (s *SessionService) GetDemo(id int) (*demorecord.DemoRecordRaw, error) {
	row := s.db.QueryRow(`SELECT data FROM session_demo WHERE session_id = ?`, id)

	var compressed []byte
	err := row.Scan(&compressed)

	if err != nil {
		return nil, err
	}

	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}

	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return demorecord.Parse(raw)
}

func (s *SessionService) GetDemoPlayers(demo *demorecord.DemoRecordParsed) {
	userIds := []int{}

	lookup := map[int]*demorecord.DemoRecordParsedPlayer{}

	for i := range demo.Players {
		item := demo.Players[i]

		if user, _ := s.usersService.GetByAuth(item.UniqueId, models.AuthType(item.UserType)); user != nil {
			userIds = append(userIds, user.Id)
			lookup[user.Id] = item
		}
	}

	if profiles, _ := s.usersService.GetUserProfiles(userIds); profiles != nil {
		for i := range profiles {
			profile := profiles[i]

			lookup[profile.Id].Profile = profile
		}
	}
}

func (s *SessionService) LoadDemoUsers(demo *demorecord.DemoRecordParsed) {
	for i := range demo.Players {
		item := demo.Players[i]

		if user, _ := s.usersService.GetByAuth(item.UniqueId, models.AuthType(item.UserType)); user != nil {
			item.Profile = &models.UserProfile{
				Id:     user.Id,
				Name:   user.Name,
				AuthId: user.AuthId,
			}
		}
	}
}

func (s *SessionService) UpdateGameData(data UpdateGameDataRequest) error {
	status, err := s.getStatus(data.SessionId)
	if err != nil || (*status != models.InProgress && *status != models.Lobby) {
		return err
	}

	gd := data.GameData

	_, err = s.db.Exec(`
		UPDATE session
		SET updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		data.SessionId,
	)

	if err != nil {
		return err
	}

	s.db.Exec(`
		UPDATE session_game_data
		SET max_players = ?, players_online = ?, players_alive = ?,
			wave = ?, is_trader_time = ?, zeds_left = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE session_id = ?`,
		gd.MaxPlayers, gd.PlayersOnline, gd.PlayersAlive,
		gd.Wave, gd.IsTraderTime, gd.ZedsLeft,
		data.SessionId,
	)

	if data.CDData != nil {
		length, err := s.getLength(data.SessionId)
		if err != nil {
			return err
		}

		if data.GameData.Wave <= *length {
			cdData := data.CDData

			s.db.Exec(`
			INSERT INTO session_game_data_extra 
				(session_id, spawn_cycle, max_monsters, wave_size_fakes, zeds_type)
			VALUES (?, ?, ?, ?, ?)
				ON DUPLICATE KEY UPDATE
				spawn_cycle = ?, max_monsters = ?, wave_size_fakes = ?, zeds_type = ?`,
				data.SessionId,
				cdData.SpawnCycle, cdData.MaxMonsters, cdData.WaveSizeFakes, cdData.ZedsType,
				cdData.SpawnCycle, cdData.MaxMonsters, cdData.WaveSizeFakes, cdData.ZedsType,
			)
		}
	}

	ids := []int{-1}
	for _, p := range data.Players {
		var id int
		err := s.db.QueryRow("SELECT id FROM users WHERE auth_id = ? AND auth_type = ?",
			p.AuthId, p.AuthType).Scan(&id)

		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}

			return err
		}

		ids = append(ids, id)
		_, err = s.db.Exec(`
				UPDATE users_activity 
				SET current_session_id = ?,
					perk = ?, level = ?, prestige = ?, 
					health = ?, armor = ?, is_spectator = ?,
					updated_at = CURRENT_TIMESTAMP
				WHERE user_id = ?`,
			data.SessionId, p.Perk, p.Level, p.Prestige,
			p.Health, p.Armor, p.IsSpectator, id,
		)
		if err != nil {
			return err
		}
	}

	_, err = s.db.Exec(fmt.Sprintf(`
			UPDATE users_activity 
			SET last_session_id = current_session_id, 
				current_session_id = NULL, 
				updated_at = CURRENT_TIMESTAMP
			WHERE current_session_id = %v AND user_id NOT IN (%v)`,
		data.SessionId, util.IntArrayToString(ids, ",")),
	)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(fmt.Sprintf(`
			UPDATE users_activity 
			SET current_session_id = %v, 
				updated_at = CURRENT_TIMESTAMP
			WHERE user_id IN (%v)`,
		data.SessionId, util.IntArrayToString(ids, ",")),
	)
	if err != nil {
		return err
	}

	return err
}

func (s *SessionService) getLength(id int) (*models.GameLength, error) {
	row := s.db.QueryRow(`SELECT length FROM session WHERE id = ?`, id)

	var length models.GameLength
	err := row.Scan(&length)

	if err != nil {
		return nil, err
	}

	return &length, nil
}

func (s *SessionService) getStatus(id int) (*models.GameStatus, error) {
	row := s.db.QueryRow(`SELECT status FROM session WHERE id = ?`, id)

	var status models.GameStatus
	err := row.Scan(&status)

	if err != nil {
		return nil, err
	}

	return &status, nil
}
