package maps

import (
	"database/sql"
)

type MapsService struct {
	db *sql.DB
}

func (s *MapsService) initTables() {
	s.db.Exec(`
	CREATE TABLE IF NOT EXISTS maps (
		id INTEGER PRIMARY KEY AUTO_INCREMENT,
		name VARCHAR(64), 
		preview TEXT,
		
		UNIQUE INDEX idx_maps_name (name)
	);
	`)
}

func NewMapsService(db *sql.DB) *MapsService {
	service := MapsService{
		db: db,
	}

	service.initTables()

	return &service
}

func (s *MapsService) Create(req AddMapRequest) (int, error) {
	_, err := s.db.Exec(`
		INSERT INTO maps (name, preview) VALUES ($1, $2)
			ON CONFLICT(name) DO UPDATE SET preview = $2`,
		req.Name, req.Preview)

	if err != nil {
		return 0, err
	}

	data, err := s.getByName(req.Name)
	if err != nil {
		return 0, err
	}

	return data.Id, err
}

func (s *MapsService) GetByPattern(pattern string) ([]Map, error) {
	rows, err := s.db.Query(`
		SELECT * FROM maps 
		WHERE (name LIKE $1)`,
		"%"+pattern+"%")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	items := []Map{}

	for rows.Next() {
		item := Map{}

		err := rows.Scan(&item.Id, &item.Name, &item.Preview)
		if err != nil {
			continue
		}

		items = append(items, item)
	}

	return items, nil
}

func (s *MapsService) GetById(id int) (*Map, error) {
	row := s.db.QueryRow(`SELECT * FROM maps WHERE id = $1`, id)

	item := Map{}

	err := row.Scan(&item.Id, &item.Name, &item.Preview)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *MapsService) getByName(name string) (*Map, error) {
	row := s.db.QueryRow(`SELECT * FROM maps WHERE name = $1`, name)

	item := Map{}

	err := row.Scan(&item.Id, &item.Name, &item.Preview)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *MapsService) UpdatePreview(data UpdatePreviewRequest) error {
	_, err := s.db.Exec(`UPDATE maps SET preview = $1 WHERE id = $2`,
		data.Preview, data.Id)

	return err
}
