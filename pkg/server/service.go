package server

import (
	"database/sql"
)

type ServerService struct {
	db *sql.DB
}

func NewServerService(db *sql.DB) *ServerService {
	service := ServerService{
		db: db,
	}

	return &service
}

func (s *ServerService) Create(req AddServerRequest) (int, error) {
	_, err := s.db.Exec(`
		INSERT INTO server (name, address) VALUES (?, ?)
			ON DUPLICATE KEY UPDATE name = ?`,
		req.Name, req.Address, req.Name)

	if err != nil {
		return 0, err
	}

	data, err := s.getByAddress(req.Address)
	if err != nil {
		return 0, err
	}

	return data.Id, err
}

func (s *ServerService) GetByPattern(pattern string) ([]Server, error) {
	sqlPattern := "%" + pattern + "%"
	rows, err := s.db.Query(`
		SELECT id, name, address FROM server 
		WHERE (address LIKE ?) OR (name LIKE ?)
		ORDER BY name`, sqlPattern, sqlPattern)
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
	row := s.db.QueryRow(`SELECT id, name, address FROM server WHERE id = ?`, id)

	server := Server{}

	err := row.Scan(&server.Id, &server.Name, &server.Address)
	if err != nil {
		return nil, err
	}

	return &server, nil
}

func (s *ServerService) getByAddress(address string) (*Server, error) {
	row := s.db.QueryRow(`SELECT id, name, address FROM server WHERE address = ?`, address)

	server := Server{}

	err := row.Scan(&server.Id, &server.Name, &server.Address)
	if err != nil {
		return nil, err
	}

	return &server, nil
}

func (s *ServerService) UpdateName(data UpdateNameRequest) error {
	_, err := s.db.Exec(`UPDATE server SET name = ? WHERE id = ?`,
		data.Name, data.Id)

	return err
}
