package users

import (
	"database/sql"
	"fmt"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/steamapi"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
)

type UserService struct {
	db *sql.DB

	steamApiService *steamapi.SteamApiUserService
}

func (s *UserService) initTables() {
	s.db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTO_INCREMENT,
		auth_id STRING NOT NULL,
		auth_type INTEGER NOT NULL,

		name STRING NOT NULL
	);

	CREATE UNIQUE INDEX IF NOT EXISTS idx_users_auth ON users (auth_id, auth_type);

	CREATE TABLE IF NOT EXISTS users_activity (
		user_id INTEGER PRIMARY KEY REFERENCES users(id)
			ON UPDATE CASCADE
			ON DELETE CASCADE,
		
		current_session_id INTEGER NULL REFERENCES session(id)
			ON UPDATE SET NULL
			ON DELETE SET NULL,

		last_session_id INTEGER NULL REFERENCES session(id)
			ON UPDATE SET NULL
			ON DELETE SET NULL,
		
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_users_activity_curr ON users_activity (current_session_id);

	CREATE TRIGGER IF NOT EXISTS update_user_activity_on_wave_end
	AFTER INSERT ON wave_stats_player
	FOR EACH ROW
	BEGIN
		UPDATE users_activity
		SET current_session_id = 
			(select min(ws.session_id) from wave_stats ws
			inner join wave_stats_player wsp on wsp.stats_id = ws.id
			where wsp.id = new.id),
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = new.player_id;
	END;

	CREATE TRIGGER IF NOT EXISTS update_user_activity_on_session_end
	AFTER UPDATE OF status ON session
	FOR EACH ROW
	WHEN new.status IN (2,3,4,-1)
	BEGIN
		UPDATE users_activity
		SET last_session_id = current_session_id, 
			current_session_id = NULL, 
			updated_at = CURRENT_TIMESTAMP
		WHERE current_session_id = new.id;
	END;
	`)
}

func NewUserService(db *sql.DB) *UserService {
	service := UserService{
		db: db,
	}

	service.initTables()

	return &service
}

func (s *UserService) Inject(
	steamApiService *steamapi.SteamApiUserService,
) {
	s.steamApiService = steamApiService
}

func (s *UserService) FindCreateFind(req CreateUserRequest) (int, error) {
	data, err := s.getByAuth(req.AuthId, req.AuthType)
	if err == nil {
		return data.Id, nil
	}

	_, err = s.db.Exec(`
		INSERT INTO users (auth_id, auth_type, name) 
		VALUES ($1, $2, $3)`,
		req.AuthId, req.AuthType, req.Name,
	)

	if err != nil {
		return 0, err
	}

	data, err = s.getByAuth(req.AuthId, req.AuthType)
	if err != nil {
		return 0, err
	}

	_, err = s.db.Exec(`
		INSERT INTO users_activity (user_id, current_session_id, last_session_id) 
		VALUES ($1, NULL, NULL)`, data.Id,
	)

	return data.Id, err
}

func (s *UserService) getByAuth(authId string, authType models.AuthType) (*User, error) {
	row := s.db.QueryRow(`
		SELECT * FROM users WHERE auth_id = $1 AND auth_type = $2`,
		authId, authType,
	)

	item := User{}

	err := row.Scan(&item.Id, &item.AuthId, &item.Type, &item.Name)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *UserService) filter(req FilterUsersRequest) (*FilterUsersResponse, error) {
	page, limit := util.ParsePagination(req.Pager)

	sql := fmt.Sprintf(`
		SELECT 
			users.id,
			users.name,
			users.auth_id,
			users.auth_type,
			users_activity.last_session_id,
			users_activity.current_session_id,
			users_activity.updated_at
		FROM users
		INNER JOIN users_activity ON users_activity.user_id = users.id
		ORDER BY users_activity.updated_at DESC
		LIMIT %v, %v
		`, page*limit, limit,
	)

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	sessionIdSet := make(map[int]bool)
	sessionIds := []int{}
	steamIdSet := make(map[string]bool)
	steamIds := []string{}
	items := []*FilterUsersResponseUser{}

	for rows.Next() {
		item := FilterUsersResponseUser{}

		rows.Scan(
			&item.Id, &item.Name, &item.AuthId, &item.Type,
			&item.LastSessionId, &item.CurrentSessionId, &item.UpdatedAt,
		)

		if item.LastSessionId != nil {
			sessionIdSet[*item.LastSessionId] = true
		}
		if item.CurrentSessionId != nil {
			sessionIdSet[*item.CurrentSessionId] = true
		}
		if item.Type == models.Steam {
			steamIdSet[item.AuthId] = true
		}

		items = append(items, &item)
	}

	for key := range sessionIdSet {
		sessionIds = append(sessionIds, key)
	}
	for key := range steamIdSet {
		steamIds = append(steamIds, key)
	}

	sessions, err := s.getSessions(sessionIds)
	steamData, err := s.getSteamData(steamIds)

	for _, item := range items {
		if item.LastSessionId != nil {
			if data, ok := sessions[*item.LastSessionId]; ok {
				item.LastSession = &data
			}
		}

		if item.CurrentSessionId != nil {
			if data, ok := sessions[*item.CurrentSessionId]; ok {
				item.CurrentSession = &data
			}
		}

		if data, ok := steamData[item.AuthId]; ok {
			item.Avatar = &data.Avatar
			item.ProfileUrl = &data.ProfileUrl
		}
	}

	return &FilterUsersResponse{
		Items: items,
	}, nil
}

func (s *UserService) getSessions(sessionIds []int) (map[int]FilterUsersResponseUserSession, error) {
	sql := fmt.Sprintf(`
		SELECT 
			session.id,
			session.mode,
			session.length,
			session.diff,
			session.status,
			cd.spawn_cycle,
			cd.max_monsters,
			cd.wave_size_fakes,
			cd.zeds_type,
			server.name,
			maps.name
		FROM session
		INNER JOIN server on server.id = session.server_id
		INNER JOIN maps on maps.id = session.map_id
		LEFT JOIN session_game_data_cd cd on cd.session_id = session.id
		WHERE session.id IN (%v)`,
		util.IntArrayToString(sessionIds, ","),
	)

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	items := make(map[int]FilterUsersResponseUserSession)
	for rows.Next() {
		item := FilterUsersResponseUserSession{}
		cdData := models.CDGameData{}

		rows.Scan(
			&item.Id, &item.Mode, &item.Length, &item.Difficulty, &item.Status,
			&cdData.SpawnCycle, &cdData.MaxMonsters, &cdData.WaveSizeFakes, &cdData.ZedsType,
			&item.ServerName, &item.MapName,
		)

		if cdData.SpawnCycle != nil {
			item.CDData = &cdData
		}

		items[item.Id] = item
	}

	return items, nil
}

func (s *UserService) getSteamData(steamIds []string) (map[string]steamapi.GetUserSummaryPlayer, error) {
	steamData, err := s.steamApiService.GetUserSummary(steamIds)
	if err != nil {
		return nil, err
	}

	steamDataSet := make(map[string]steamapi.GetUserSummaryPlayer)
	for _, data := range steamData {
		steamDataSet[data.SteamId] = data
	}

	return steamDataSet, nil
}
