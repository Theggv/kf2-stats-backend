package stats

import (
	"database/sql"

	"github.com/theggv/kf2-stats-backend/pkg/users"
)

type StatsService struct {
	db          *sql.DB
	userService *users.UserService
}

func (s *StatsService) Inject(userService *users.UserService) {
	s.userService = userService
}

func (s *StatsService) initTables() {
	s.db.Exec(`
	CREATE TABLE IF NOT EXISTS wave_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER NOT NULL REFERENCES session(id) ON UPDATE CASCADE,
		player_id INTEGER NOT NULL REFERENCES user(id) ON UPDATE CASCADE,
		wave INTEGER NOT NULL,
		attempt INTEGER NOT NULL,

		perk INTEGER NOT NULL,

		shots_fired INTEGER NOT NULL,
		shots_hit INTEGER NOT NULL,
		shots_hs INTEGER NOT NULL,

		dosh_earned INTEGER NOT NULL,

		heals_given INTEGER NOT NULL,
		heals_recv INTEGER NOT NULL,

		damage_dealt INTEGER NOT NULL,
		damage_taken INTEGER NOT NULL,

		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS wave_stats_kills (
		stats_id INTEGER PRIMARY KEY REFERENCES wave_stats(id) ON UPDATE CASCADE,

		cyst INTEGER NOT NULL,
		alpha_clot INTEGER NOT NULL,
		slasher INTEGER NOT NULL,
		stalker INTEGER NOT NULL,
		crawler INTEGER NOT NULL,
		gorefast INTEGER NOT NULL,
		rioter INTEGER NOT NULL,
		elite_crawler INTEGER NOT NULL,
		gorefiend INTEGER NOT NULL,

		siren INTEGER NOT NULL,
		bloat INTEGER NOT NULL,
		edar INTEGER NOT NULL,
		husk_n INTEGER NOT NULL,
		husk_b INTEGER NOT NULL,
		husk_r INTEGER NOT NULL,

		scrake INTEGER NOT NULL,
		fp INTEGER NOT NULL,
		qp INTEGER NOT NULL,
		boss INTEGER NOT NULL
	);
	
	CREATE UNIQUE INDEX IF NOT EXISTS idx_wave_stats ON wave_stats (
		session_id, player_id, wave, attempt
	);`)
}

func NewStatsService(db *sql.DB) *StatsService {
	service := StatsService{
		db: db,
	}

	service.initTables()

	return &service
}

func (s *StatsService) CreateWaveStats(req CreateWaveStatsRequest) error {
	playerId, err := s.userService.FindCreateFind(users.CreateUserRequest{
		AuthId: req.UserAuthId,
		Type:   req.UserAuthType,
		Name:   req.UserName,
	})

	if err != nil {
		return err
	}

	row := s.db.QueryRow(
		`SELECT COUNT(*) FROM wave_stats
		WHERE session_id = $1 AND player_id = $2 AND wave = $3`,
		req.SessionId, playerId, req.Wave,
	)

	var attempt int
	err = row.Scan(&attempt)

	res, err := s.db.Exec(`
		INSERT INTO wave_stats (
			session_id, player_id, wave, attempt, 
			perk, shots_fired, shots_hit, shots_hs, 
			dosh_earned, heals_given, heals_recv,
			damage_dealt, damage_taken) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		req.SessionId, playerId, req.Wave, attempt+1,
		req.Perk, req.ShotsFired, req.ShotsHit, req.ShotsHS,
		req.DoshEarned, req.HealsGiven, req.HealsReceived,
		req.DamageDealt, req.DamageTaken,
	)

	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	kills := req.Kills

	_, err = s.db.Exec(`
		INSERT INTO wave_stats_kills (stats_id, 
			cyst, alpha_clot, slasher, stalker, crawler, gorefast, 
			rioter, elite_crawler, gorefiend, 
			siren, bloat, edar, 
			husk_n, husk_b, husk_r, 
			scrake, fp, qp, boss) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`,
		int(id),
		kills.Cyst, kills.AlphaClot, kills.Slasher, kills.Stalker, kills.Crawler, kills.Gorefast,
		kills.Rioter, kills.EliteCrawler, kills.Gorefiend,
		kills.Siren, kills.Bloat, kills.Edar,
		kills.Husk, req.HuskBackpackKills, req.HuskRages,
		kills.Scrake, kills.FP, kills.QP, kills.Boss,
	)

	return err
}
