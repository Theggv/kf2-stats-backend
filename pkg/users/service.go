package users

import (
	"database/sql"
)

type UserService struct {
	db *sql.DB
}

func (s *UserService) initTables() {
	s.db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		auth_id STRING NOT NULL,
		auth_type INTEGER NOT NULL,

		name STRING NOT NULL
	);

	CREATE TABLE IF NOT EXISTS users_name_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER REFERENCES users(id) ON UPDATE CASCADE,
		name STRING NOT NULL,

		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE UNIQUE INDEX IF NOT EXISTS idx_users_auth ON users (auth_id, auth_type);
	`)
}

func NewUserService(db *sql.DB) *UserService {
	service := UserService{
		db: db,
	}

	service.initTables()

	return &service
}

func (s *UserService) FindCreateFind(req CreateUserRequest) (int, error) {
	data, err := s.getByAuth(req.AuthId, req.Type)
	if err == nil {
		return data.Id, nil
	}

	_, err = s.db.Exec(`
		INSERT INTO users (auth_id, auth_type, name) 
		VALUES ($1, $2, $3)`,
		req.AuthId, req.Type, req.Name,
	)

	if err != nil {
		return 0, err
	}

	data, err = s.getByAuth(req.AuthId, req.Type)

	return data.Id, err
}

func (s *UserService) getByAuth(authId string, authType AuthType) (*User, error) {
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
