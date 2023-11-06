package maps

import (
	"database/sql"
)

type MapsService struct {
	db *sql.DB
}

func NewMapsService(db *sql.DB) *MapsService {
	service := MapsService{
		db: db,
	}

	return &service
}

func (s *MapsService) Create(req AddMapRequest) (int, error) {
	_, err := s.db.Exec(`
		INSERT INTO maps (name, preview) VALUES (?, ?)
			ON DUPLICATE KEY UPDATE preview = ?`,
		req.Name, req.Preview, req.Preview)

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
		SELECT id, name, preview FROM maps 
		WHERE (name LIKE ?)`,
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
	row := s.db.QueryRow(`SELECT id, name, preview FROM maps WHERE id = ?`, id)

	item := Map{}

	err := row.Scan(&item.Id, &item.Name, &item.Preview)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *MapsService) getByName(name string) (*Map, error) {
	row := s.db.QueryRow(`SELECT id, name, preview FROM maps WHERE name = ?`, name)

	item := Map{}

	err := row.Scan(&item.Id, &item.Name, &item.Preview)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *MapsService) UpdatePreview(data UpdatePreviewRequest) error {
	_, err := s.db.Exec(`UPDATE maps SET preview = ? WHERE ? = $2`,
		data.Preview, data.Id)

	return err
}
