package stats

import (
	"database/sql"
)

type StatsService struct {
	db *sql.DB
}

func (s *StatsService) initTables() {
	s.db.Exec(`
	CREATE TABLE IF NOT EXISTS stats (
		session_id INTEGER NOT NULL REFERENCES session(id) ON UPDATE CASCADE,
		player_id INTEGER NOT NULL REFERENCES user(id) ON UPDATE CASCADE,
		wave INTEGER NOT NULL,
		attempt INTEGER NOT NULL,

		perk INTEGER NOT NULL,

		acc REAL NOT NULL,
		hs_acc REAL NOT NULL,

		trash_kills INTEGER NOT NULL,
		medium_kills INTEGER NOT NULL,
		scrake_kills INTEGER NOT NULL,
		fp_kills INTEGER NOT NULL,
		minifp_kills INTEGER NOT NULL,
		boss_kills INTEGER NOT NULL,

		husk_n INTEGER NOT NULL,
		husk_b INTEGER NOT NULL,
		husk_r INTEGER NOT NULL,

		damage_dealt INTEGER NOT NULL,
		damage_taken INTEGER NOT NULL,

		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

		PRIMARY KEY (session_id, player_id, wave, attempt)
	);`)
}

func NewStatsService(db *sql.DB) *StatsService {
	service := StatsService{
		db: db,
	}

	service.initTables()

	return &service
}

func (s *StatsService) CreateStats(req CreateStatsRequest) error {
	_, err := s.db.Exec(`
		INSERT INTO stats (
			session_id, player_id, wave, attempt, 
			perk, acc, hs_acc, 
			trash_kills, medium_kills, scrake_kills, 
			fp_kills, minifp_kills, boss_kills, 
			husk_n, husk_b, husk_r, 
			damage_dealt, damage_taken) 
		VALUES ($1, $2, $3, $4, $5)`,
		req.SessionId, req.PlayerId, req.Wave, req.Attempt,
		req.Perk, req.Accuracy, req.HSAccuracy,
		req.TrashKills, req.MediumKills, req.ScrakeKills,
		req.FPKills, req.MiniFPKills, req.BossKills,
		req.HuskNormalKills, req.HuskBackpackKills, req.HuskRages,
		req.DamageDealt, req.DamageTaken,
	)

	return err
}
