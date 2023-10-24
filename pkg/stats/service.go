package stats

import (
	"database/sql"
	"fmt"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
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
		session_id INTEGER NOT NULL REFERENCES session(id)
			ON UPDATE CASCADE 
			ON DELETE CASCADE,
		wave INTEGER NOT NULL,
		attempt INTEGER NOT NULL,

		started_at DATETIME NOT NULL,
		completed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE UNIQUE INDEX IF NOT EXISTS uniq_wave_stats ON wave_stats (
		session_id, wave, attempt
	);

	CREATE TABLE IF NOT EXISTS wave_stats_player (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		stats_id INTEGER NOT NULL REFERENCES wave_stats(id)
			ON UPDATE CASCADE 
			ON DELETE CASCADE,
		player_id INTEGER NOT NULL REFERENCES user(id)
			ON UPDATE CASCADE 
			ON DELETE CASCADE,

		perk INTEGER NOT NULL,
		level INTEGER NOT NULL,
		prestige INTEGER NOT NULL,

		is_dead BOOLEAN NOT NULL,

		shots_fired INTEGER NOT NULL,
		shots_hit INTEGER NOT NULL,
		shots_hs INTEGER NOT NULL,

		dosh_earned INTEGER NOT NULL,

		heals_given INTEGER NOT NULL,
		heals_recv INTEGER NOT NULL,

		damage_dealt INTEGER NOT NULL,
		damage_taken INTEGER NOT NULL,

		zedtime_count INTEGER NOT NULL,
		zedtime_length REAL NOT NULL,

		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE UNIQUE INDEX IF NOT EXISTS uniq_wave_stats_player ON wave_stats_player (
		stats_id, player_id
	);

	CREATE INDEX IF NOT EXISTS idx_wave_stats_player_player_id ON wave_stats_player (
		player_id
	);

	CREATE TABLE IF NOT EXISTS wave_stats_player_kills (
		player_stats_id INTEGER PRIMARY KEY REFERENCES wave_stats_player(id) 
			ON UPDATE CASCADE 
			ON DELETE CASCADE,

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

	CREATE TABLE IF NOT EXISTS wave_stats_player_injured_by (
		player_stats_id INTEGER PRIMARY KEY REFERENCES wave_stats_player(id) 
			ON UPDATE CASCADE 
			ON DELETE CASCADE,

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
		husk INTEGER NOT NULL,

		scrake INTEGER NOT NULL,
		fp INTEGER NOT NULL,
		qp INTEGER NOT NULL,
		boss INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS wave_stats_cd (
		stats_id INTEGER PRIMARY KEY REFERENCES wave_stats(id)
			ON UPDATE CASCADE 
			ON DELETE CASCADE,

		spawn_cycle TEXT NOT NULL,
		max_monsters INTEGER NOT NULL,
		wave_size_fakes INTEGER NOT NULL,
		zeds_type TEXT NOT NULL,

		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS aggregated_kills (
		player_stats_id INTEGER PRIMARY KEY REFERENCES wave_stats_player(id) 
			ON UPDATE CASCADE 
			ON DELETE CASCADE,

		trash INTEGER NOT NULL,
		medium INTEGER NOT NULL,
		large INTEGER NOT NULL,
		total INTEGER NOT NULL
	);

	CREATE TRIGGER insert_aggregated_kills
	AFTER INSERT ON wave_stats_player_kills
	BEGIN
		INSERT INTO aggregated_kills (player_stats_id, trash, medium, large, total)
		VALUES (
			new.player_stats_id,
			new.cyst + new.alpha_clot + new.slasher + 
			new.stalker + new.crawler + new.gorefast + 
			new.rioter + new.elite_crawler + new.gorefiend,
			new.siren + new.bloat + new.edar + new.husk_n + new.husk_b, 
			new.scrake + new.fp + new.qp, 
			new.cyst + new.alpha_clot + new.slasher + 
			new.stalker + new.crawler + new.gorefast + 
			new.rioter + new.elite_crawler + new.gorefiend + 
			new.siren + new.bloat + new.edar + new.husk_n + new.husk_b + 
			new.scrake + new.fp + new.qp + new.boss
		);
	END;
	`)
}

func NewStatsService(db *sql.DB) *StatsService {
	service := StatsService{
		db: db,
	}

	service.initTables()

	return &service
}

func (s *StatsService) getWaveAttempts(sessionId, wave int) (int, error) {
	row := s.db.QueryRow(`
		SELECT COUNT(*) FROM wave_stats
		WHERE session_id = $1 AND wave = $2`,
		sessionId, wave,
	)

	var attempt int
	err := row.Scan(&attempt)

	return attempt, err
}

func (s *StatsService) createWaveStats(req *CreateWaveStatsRequest) (int64, error) {
	attempt, err := s.getWaveAttempts(req.SessionId, req.Wave)
	if err != nil {
		return 0, err
	}

	sql := fmt.Sprintf(` 
		INSERT INTO wave_stats (session_id, wave, attempt, started_at) 
		VALUES (%v, %v, %v, datetime(CURRENT_TIMESTAMP, '-%v seconds'))`,
		req.SessionId, req.Wave, attempt+1, req.Length,
	)

	res, err := s.db.Exec(sql)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	return id, err
}

func (s *StatsService) createWaveStatsPlayer(statsId int, req *CreateWaveStatsRequestPlayer) error {
	playerId, err := s.userService.FindCreateFind(users.CreateUserRequest{
		AuthId: req.UserAuthId,
		Type:   req.UserAuthType,
		Name:   req.UserName,
	})

	if err != nil {
		return err
	}

	res, err := s.db.Exec(`
		INSERT INTO wave_stats_player (
			stats_id, player_id, 
			perk, level, prestige, is_dead,
			shots_fired, shots_hit, shots_hs, 
			dosh_earned, heals_given, heals_recv,
			damage_dealt, damage_taken,
			zedtime_count, zedtime_length) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
		statsId, playerId,
		req.Perk, req.Level, req.Prestige, req.IsDead,
		req.ShotsFired, req.ShotsHit, req.ShotsHS,
		req.DoshEarned, req.HealsGiven, req.HealsReceived,
		req.DamageDealt, req.DamageTaken,
		req.ZedTimeCount, req.ZedTimeLength,
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
		INSERT INTO wave_stats_player_kills (player_stats_id, 
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

	injuredby := req.Injuredby

	_, err = s.db.Exec(`
		INSERT INTO wave_stats_player_injured_by (player_stats_id, 
			cyst, alpha_clot, slasher, stalker, crawler, gorefast, 
			rioter, elite_crawler, gorefiend, 
			siren, bloat, edar, husk, 
			scrake, fp, qp, boss) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
		int(id),
		injuredby.Cyst, injuredby.AlphaClot, injuredby.Slasher, injuredby.Stalker, injuredby.Crawler, injuredby.Gorefast,
		injuredby.Rioter, injuredby.EliteCrawler, injuredby.Gorefiend,
		injuredby.Siren, injuredby.Bloat, injuredby.Edar, injuredby.Husk,
		injuredby.Scrake, injuredby.FP, injuredby.QP, injuredby.Boss,
	)

	return err
}

func (s *StatsService) createWaveStatsCD(statsId int, req *models.CDGameData) error {
	_, err := s.db.Exec(`
		INSERT INTO wave_stats_cd (
			stats_id, 
			spawn_cycle, max_monsters, wave_size_fakes, zeds_type) 
		VALUES ($1, $2, $3, $4, $5)`,
		statsId,
		req.SpawnCycle, req.MaxMonsters,
		req.WaveSizeFakes, req.ZedsType,
	)

	return err
}

func (s *StatsService) CreateWaveStats(req CreateWaveStatsRequest) error {
	statsId, err := s.createWaveStats(&req)
	if err != nil {
		return err
	}

	for _, player := range req.Players {
		// Skip players without stats
		if player.Perk == 0 && player.Level == 0 && player.DamageDealt == 0 && player.DamageTaken == 0 {
			continue
		}

		err = s.createWaveStatsPlayer(int(statsId), &player)
		if err != nil {
			return err
		}
	}

	if req.CDData != nil && req.CDData.SpawnCycle != nil {
		err = s.createWaveStatsCD(int(statsId), req.CDData)
	}

	return err
}
