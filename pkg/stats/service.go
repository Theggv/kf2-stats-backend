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

func NewStatsService(db *sql.DB) *StatsService {
	service := StatsService{
		db: db,
	}

	return &service
}

func (s *StatsService) getWaveAttempts(sessionId, wave int) (int, error) {
	row := s.db.QueryRow(`
		SELECT COUNT(*) FROM wave_stats
		WHERE session_id = ? AND wave = ?`,
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
		VALUES (%v, %v, %v, TIMESTAMPADD(SECOND, -%v, CURRENT_TIMESTAMP))`,
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
		AuthId:   req.UserAuthId,
		AuthType: req.UserAuthType,
		Name:     req.UserName,
	})

	if err != nil {
		return err
	}

	if req.ShotsFired < 0 ||
		req.DamageDealt == 0 && req.DamageTaken == 0 && req.HealsGiven == 0 && req.HealsReceived == 0 {
		return nil
	}

	res, err := s.db.Exec(`
		INSERT INTO wave_stats_player (
			stats_id, player_id, 
			perk, level, prestige, is_dead,
			shots_fired, shots_hit, shots_hs, 
			dosh_earned, heals_given, heals_recv,
			damage_dealt, damage_taken,
			zedtime_count, zedtime_length) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
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
			scrake, fp, qp, boss, custom) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		int(id),
		kills.Cyst, kills.AlphaClot, kills.Slasher, kills.Stalker, kills.Crawler, kills.Gorefast,
		kills.Rioter, kills.EliteCrawler, kills.Gorefiend,
		kills.Siren, kills.Bloat, kills.Edar,
		kills.Husk, req.HuskBackpackKills, req.HuskRages,
		kills.Scrake, kills.FP, kills.QP, kills.Boss, kills.Custom,
	)

	injuredby := req.Injuredby

	_, err = s.db.Exec(`
		INSERT INTO wave_stats_player_injured_by (player_stats_id, 
			cyst, alpha_clot, slasher, stalker, crawler, gorefast, 
			rioter, elite_crawler, gorefiend, 
			siren, bloat, edar, husk, 
			scrake, fp, qp, boss) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		int(id),
		injuredby.Cyst, injuredby.AlphaClot, injuredby.Slasher, injuredby.Stalker, injuredby.Crawler, injuredby.Gorefast,
		injuredby.Rioter, injuredby.EliteCrawler, injuredby.Gorefiend,
		injuredby.Siren, injuredby.Bloat, injuredby.Edar, injuredby.Husk,
		injuredby.Scrake, injuredby.FP, injuredby.QP, injuredby.Boss,
	)

	_, err = s.db.Exec(`
		INSERT INTO wave_stats_player_comms (player_stats_id,
			request_healing, request_dosh, request_help, 
			taunt_zeds, follow_me, get_to_the_trader, 
			affirmative, negative, thank_you) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		int(id),
		req.RequestHealing, req.RequestDosh, req.RequestHelp,
		req.TauntZeds, req.FollowMe, req.GetToTheTrader,
		req.Affirmative, req.Negative, req.ThankYou,
	)

	return err
}

func (s *StatsService) createWaveStatsCD(statsId int, req *models.CDGameData) error {
	_, err := s.db.Exec(`
		INSERT INTO wave_stats_cd (
			stats_id, spawn_cycle, max_monsters, wave_size_fakes, zeds_type) 
		VALUES (?, ?, ?, ?, ?)`,
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
