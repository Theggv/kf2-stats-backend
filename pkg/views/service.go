package views

import "database/sql"

type ViewsService struct {
	db *sql.DB
}

func NewViewsService(db *sql.DB) *ViewsService {
	service := ViewsService{
		db: db,
	}

	return &service
}
