package session

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

type SessionService struct {
	db *sql.DB
}

func (s *SessionService) initTables() {
	s.db.Exec(`
	CREATE TABLE IF NOT EXISTS session (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER NOT NULL REFERENCES server(id) ON UPDATE CASCADE,
		map_id INTEGER NOT NULL,

		mode INTEGER NOT NULL,
		length INTEGER NOT NULL,
		diff INTEGER NOT NULL,

		status INTEGER NOT NULL DEFAULT 1,

		created_at DATETIME DEFAULT CURRENT_TIMESTAMP, 
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`)
}

func NewSessionService(db *sql.DB) *SessionService {
	service := SessionService{
		db: db,
	}

	service.initTables()

	return &service
}

func (s *SessionService) CreateSession(req CreateSessionRequest) (int, error) {
	res, err := s.db.Exec(`
		INSERT INTO session (server_id, map_id, mode, length, diff) 
		VALUES ($1, $2, $3, $4, $5)`,
		req.ServerId, req.MapId, req.Mode, req.Length, req.Difficulty)

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()

	return int(id), err
}

func (s *SessionService) FilterSessions(req FilterSessionsRequest) (*FilterSessionsResponse, error) {
	page, limit := parsePagination(req.Pager)

	// Prepare filter query
	conditions := []string{}
	conditions = append(conditions, "1") // in case if no filters passed

	if len(req.ServerId) > 0 {
		conditions = append(conditions,
			fmt.Sprintf("server_id in (%s)", intArrayToString(req.ServerId, ",")),
		)
	}

	if len(req.MapId) > 0 {
		conditions = append(conditions,
			fmt.Sprintf("map_id in (%s)", intArrayToString(req.MapId, ",")),
		)
	}

	if req.Difficulty != 0 {
		conditions = append(conditions, fmt.Sprintf("diff = %v", req.Difficulty))
	}

	if req.Length != 0 {
		conditions = append(conditions, fmt.Sprintf("length = %v", req.Length))
	}

	if req.Mode != 0 {
		conditions = append(conditions, fmt.Sprintf("mode = %v", req.Mode))
	}

	sql := fmt.Sprintf(`
		SELECT * FROM session
		WHERE %v
		LIMIT %v, %v`,
		strings.Join(conditions, " AND "), page*limit, limit,
	)

	// Execute filter query
	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	items := []Session{}

	// Parsing results
	for rows.Next() {
		item := Session{}

		err := rows.Scan(
			&item.Id, &item.ServerId, &item.MapId,
			&item.Mode, &item.Length, &item.Difficulty,
			&item.Status, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			fmt.Print(err)
			continue
		}

		items = append(items, item)
	}

	// Prepare count query
	sql = fmt.Sprintf(`
		SELECT COUNT(*) FROM session
		WHERE %v`,
		strings.Join(conditions, " AND "),
	)

	// Execute count query
	row := s.db.QueryRow(sql)

	// Parsing results
	var total int
	if row.Scan(&total) != nil {
		return nil, err
	}

	return &FilterSessionsResponse{
		Items: items,
		Metadata: models.PaginationResponse{
			Page:           page,
			ResultsPerPage: limit,
			TotalResults:   total,
		},
	}, nil
}

func (s *SessionService) GetById(id int) (*Session, error) {
	row := s.db.QueryRow(`SELECT * FROM session WHERE id = $1`, id)

	item := Session{}

	err := row.Scan(
		&item.Id, &item.ServerId, &item.MapId,
		&item.Mode, &item.Length, &item.Difficulty,
		&item.Status, &item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *SessionService) UpdateStatus(data UpdateStatusRequest) error {
	_, err := s.db.Exec(`
		UPDATE session 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2`,
		data.Status, data.Id)

	return err
}

func parsePagination(pager models.PaginationRequest) (int, int) {
	page := pager.Page
	resultsPerPage := pager.ResultsPerPage

	if page < 0 {
		page = 0
	}

	if resultsPerPage < 10 {
		resultsPerPage = 10
	}

	if resultsPerPage > 100 {
		resultsPerPage = 100
	}

	return page, resultsPerPage
}

func intArrayToString(a []int, delimiter string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delimiter, -1), "[]")
}
