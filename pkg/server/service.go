package server

import (
	"database/sql"
)

type ServerService struct {
	db *sql.DB
}

func (s *ServerService) initTables() {
	s.db.Exec(`
	CREATE TABLE IF NOT EXISTS server (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT, 
		address TEXT,
		type INTEGER 
	);
	
	CREATE UNIQUE INDEX IF NOT EXISTS idx_server_address ON server (address);
	`)
}

func NewServerService(db *sql.DB) *ServerService {
	service := ServerService{
		db: db,
	}

	service.initTables()

	return &service
}

func (s *ServerService) CreateServer(req AddServerRequest) (int, error) {
	res, err := s.db.Exec(`INSERT INTO server (name, address, type) VALUES ($1, $2, $3)`,
		req.Name, req.Address, req.Type)

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()

	return int(id), err
}

func (s *ServerService) GetByPattern(pattern string) ([]Server, error) {
	rows, err := s.db.Query(`
		SELECT * FROM server 
		WHERE (address LIKE $1) OR (name LIKE $1)`,
		"%"+pattern+"%")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	results := []Server{}

	for rows.Next() {
		server := Server{}

		err := rows.Scan(&server.Id, &server.Name, &server.Address, &server.Type)
		if err != nil {
			continue
		}

		results = append(results, server)
	}

	return results, nil
}

func (s *ServerService) GetById(id int) (*Server, error) {
	row := s.db.QueryRow(`SELECT * FROM server WHERE id = $1`, id)

	server := Server{}

	err := row.Scan(&server.Id, &server.Name, &server.Address, &server.Type)
	if err != nil {
		return nil, err
	}

	return &server, nil
}
