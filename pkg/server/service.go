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
		address TEXT
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

func (s *ServerService) FindCreateFind(req AddServerRequest) (int, error) {
	data, err := s.getByAddress(req.Address)
	if err == nil {
		return data.Id, nil
	}

	_, err = s.db.Exec(`
		INSERT INTO server (name, address) VALUES ($1, $2)
			ON CONFLICT(address) DO UPDATE SET name = $1`,
		req.Name, req.Address)

	if err != nil {
		return 0, err
	}

	data, err = s.getByAddress(req.Name)

	return data.Id, err
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

		err := rows.Scan(&server.Id, &server.Name, &server.Address)
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

	err := row.Scan(&server.Id, &server.Name, &server.Address)
	if err != nil {
		return nil, err
	}

	return &server, nil
}

func (s *ServerService) getByAddress(address string) (*Server, error) {
	row := s.db.QueryRow(`SELECT * FROM server WHERE address = $1`, address)

	server := Server{}

	err := row.Scan(&server.Id, &server.Name, &server.Address)
	if err != nil {
		return nil, err
	}

	return &server, nil
}

func (s *ServerService) UpdateName(data UpdateNameRequest) error {
	_, err := s.db.Exec(`UPDATE server SET name = $1 WHERE id = $2`,
		data.Name, data.Id)

	return err
}
