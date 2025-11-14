package difficulty

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
)

type DifficultyCalculatorService struct {
	db *sql.DB

	queue map[int]bool
	mu    sync.Mutex
}

func NewDifficultyCalculator(db *sql.DB) *DifficultyCalculatorService {
	service := DifficultyCalculatorService{
		db:    db,
		queue: map[int]bool{},
	}

	go service.initQueue(30 * time.Second)

	return &service
}

func (s *DifficultyCalculatorService) AddToQueue(sessionId int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue[sessionId] = true
}

func (s *DifficultyCalculatorService) CheckIfQueued(sessionId int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.queue[sessionId]

	return exists
}

func (s *DifficultyCalculatorService) initQueue(updateTime time.Duration) {
	for range time.Tick(updateTime) {
		s.processQueue()
	}
}

func (s *DifficultyCalculatorService) processQueue() {
	items := []int{}

	s.mu.Lock()
	for item := range s.queue {
		items = append(items, item)
	}
	clear(s.queue)
	s.mu.Unlock()

	if len(items) == 0 {
		return
	}

	_, err := s.BatchRecalculate(items)
	if err != nil {
		fmt.Printf("[processQueue] %v\n", err)
	}
}

func (s *DifficultyCalculatorService) GetById(sessionId int) (*GetSessionDifficultyResponse, error) {
	items, err := s.GetByIds([]int{sessionId})
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, nil
	}

	return items[0], nil
}

func (s *DifficultyCalculatorService) GetByIds(sessionIds []int) ([]*GetSessionDifficultyResponse, error) {
	if len(sessionIds) == 0 {
		return []*GetSessionDifficultyResponse{}, nil
	}

	lookup := map[int]*GetSessionDifficultyResponse{}

	{
		stmt := fmt.Sprintf(`
			SELECT
				session_id,
				avg_zeds_diff, map_bonus,
				completion_p, restarts_penalty,
				potential_score, final_score
			FROM session_diff
			WHERE session_id IN (%v)
			`, util.IntArrayToString(sessionIds, ","),
		)

		rows, err := s.db.Query(stmt)
		if err != nil {
			return nil, err
		}

		defer rows.Close()
		for rows.Next() {
			item := GetSessionDifficultyResponse{
				Summary: &GetSessionDifficultyResponseSummary{},
				Waves:   []*GetSessionDifficultyResponseWave{},
			}

			err := rows.Scan(
				&item.SessionId,
				&item.Summary.AvgZedsDifficulty, &item.Summary.MapBonus,
				&item.Summary.CompletionPercent, &item.Summary.RestartsPenalty,
				&item.Summary.PotentialScore, &item.Summary.FinalScore,
			)
			if err != nil {
				return nil, err
			}

			lookup[item.SessionId] = &item
		}
	}

	{
		stmt := fmt.Sprintf(`
			SELECT
				session.id as session_id,
				ws.id as wave_id,
				zeds_diff,
				duration,
				predicted_duration,
				duration - predicted_duration as predicted_duration_err,
				kiting_penalty,
				wave_size_penalty,
				total_players_penalty,
				score
			FROM session
			INNER JOIN wave_stats ws
				ON ws.session_id = session.id
			INNER JOIN wave_stats_diff diff 
				ON diff.stats_id = ws.id
			WHERE 
				session.id IN (%v) AND ws.wave <= session.length
			`, util.IntArrayToString(sessionIds, ","),
		)

		rows, err := s.db.Query(stmt)
		if err != nil {
			return nil, err
		}

		defer rows.Close()
		for rows.Next() {
			var sessionId int
			item := GetSessionDifficultyResponseWave{}

			err = rows.Scan(
				&sessionId, &item.WaveId,
				&item.ZedsDifficulty, &item.Duration,
				&item.PredictedDuration, &item.PredictedDurationError,
				&item.KitingPenalty, &item.WaveSizePenalty,
				&item.TotalPlayersPenalty, &item.Score,
			)

			if err != nil {
				return nil, err
			}

			lookup[sessionId].Waves = append(lookup[sessionId].Waves, &item)
		}
	}

	res := []*GetSessionDifficultyResponse{}
	for _, value := range lookup {
		res = append(res, value)
	}

	return res, nil
}

func (s *DifficultyCalculatorService) RecalculateAll() error {
	start := time.Now()
	defer func() {
		fmt.Printf("[DifficultyCalculatorService] Completed in %v\n", time.Since(start))
	}()

	fmt.Printf("[DifficultyCalculatorService] Recalculating ALL server data.\n")

	stmt := `SELECT id FROM server`

	rows, err := s.db.Query(stmt)
	if err != nil {
		return err
	}

	serverIds := []int{}
	for rows.Next() {
		var serverId int
		err = rows.Scan(&serverId)
		if err != nil {
			return err
		}

		serverIds = append(serverIds, serverId)
	}

	for index, serverId := range serverIds {
		fmt.Printf("[DifficultyCalculatorService] server_id=%v (%v/%v)...", serverId, index+1, len(serverIds))

		now := time.Now()
		err = s.RecalculateByServerId(serverId)
		if err != nil {
			fmt.Print("failed\n")
			return err
		}

		fmt.Printf("ok (%v)\n", time.Since(now))
	}

	return nil
}

func (s *DifficultyCalculatorService) RecalculateByServerId(serverId int) error {
	stmt := `SELECT count(*) FROM session WHERE server_id = ?`

	row := s.db.QueryRow(stmt, serverId)

	var totalSessions int
	err := row.Scan(&totalSessions)
	if err != nil {
		return err
	}

	limit := 1000

	for cursor := 0; cursor < totalSessions; cursor += limit {
		stmt := `SELECT id FROM session WHERE server_id = ? LIMIT ?, ?`

		rows, err := s.db.Query(stmt, serverId, cursor, limit)
		if err != nil {
			return err
		}

		sessionIds := []int{}
		for rows.Next() {
			var sessionId int
			err = rows.Scan(&sessionId)
			if err != nil {
				return err
			}

			sessionIds = append(sessionIds, sessionId)
		}

		_, err = s.BatchRecalculate(sessionIds)
		if err != nil {
			return err
		}
	}

	return nil
}

// Batch recalculate and update session difficulties
func (s *DifficultyCalculatorService) BatchRecalculate(sessionIds []int) ([]*DifficultyCalculatorGame, error) {
	items, err := s.getSessions(sessionIds)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return items, nil
	}

	s.processItems(items, 10)

	err = util.Transact(s.db, s.updateItems(items))
	if err != nil {
		return nil, err
	}

	return items, nil
}

// Batch calculate difficulties with some concurrency level
func (s *DifficultyCalculatorService) processItems(
	items []*DifficultyCalculatorGame, maxConcurrency int,
) {
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for _, item := range items {
		wg.Add(1)
		sem <- struct{}{}
		go func(item *DifficultyCalculatorGame) {
			defer wg.Done()
			defer func() { <-sem }()
			s.processSession(item)
		}(item)
	}

	wg.Wait()
}

// Batch update session difficulties using transaction
func (s *DifficultyCalculatorService) updateItems(
	items []*DifficultyCalculatorGame,
) func(tx *sql.Tx) error {
	return func(tx *sql.Tx) error {
		{
			values := []string{}

			for _, item := range items {
				for _, wave := range item.Waves {
					res := wave.Result

					if res == nil {
						continue
					}

					values = append(values,
						fmt.Sprintf("(%v, %v, %v, %v, %v, %v, %v, %v)",
							wave.Id, res.ZedsDifficulty, wave.Duration, res.PredictedDuration,
							res.KitingPenalty, res.WaveSizePenalty, res.TotalPlayersPenalty, res.Score,
						),
					)
				}
			}

			stmt := fmt.Sprintf(`
				INSERT INTO wave_stats_diff (
					stats_id, zeds_diff, duration, predicted_duration,
					kiting_penalty, wave_size_penalty, total_players_penalty, score
				) 
				VALUES %v
				ON DUPLICATE KEY UPDATE
					zeds_diff = VALUES(zeds_diff),
					duration = VALUES(duration),
					predicted_duration = VALUES(predicted_duration),
					kiting_penalty = VALUES(kiting_penalty),
					wave_size_penalty = VALUES(wave_size_penalty),
					total_players_penalty = VALUES(total_players_penalty),
					score = VALUES(score)
				`, strings.Join(values, ","),
			)

			_, err := tx.Exec(stmt)
			if err != nil {
				return err
			}
		}

		{
			values := []string{}

			for _, item := range items {
				res := item.Result

				if res == nil {
					continue
				}

				values = append(values,
					fmt.Sprintf("(%v, %v, %v, %v, %v, %v, %v, %v)",
						item.Session.Id,
						res.AvgZedsDifficulty, res.StdDevZedsDifficulty, res.MaxZedsDifficulty,
						res.CompletionPercent, res.RestartsPenalty, res.PotentialScore, res.FinalScore,
					),
				)
			}

			stmt := fmt.Sprintf(`
				INSERT INTO session_diff (
					session_id, avg_zeds_diff, stddev_zeds_diff, max_zeds_diff,
					completion_p, restarts_penalty, potential_score, final_score
				) 
				VALUES %v
				ON DUPLICATE KEY UPDATE
					avg_zeds_diff = VALUES(avg_zeds_diff),
					stddev_zeds_diff = VALUES(stddev_zeds_diff),
					max_zeds_diff = VALUES(max_zeds_diff),
					completion_p = VALUES(completion_p),
					restarts_penalty = VALUES(restarts_penalty),
					potential_score = VALUES(potential_score),
					final_score = VALUES(final_score)
				`, strings.Join(values, ","),
			)

			_, err := tx.Exec(stmt)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// Calculate session difficulty
func (s *DifficultyCalculatorService) processSession(data *DifficultyCalculatorGame) {
	session := data.Session

	for _, wave := range data.Waves {
		zeds := wave.Zeds.ToMap()
		totalZeds := zeds.GetTotal()

		res := DifficultyCalculatorGameWaveScore{
			ZedsDifficulty: calcWaveZedsDifficulty(
				wave.ZedsType, wave.Wave, session.Length, session.Difficulty, zeds,
			),
			WaveSizePenalty:     calcWaveSizePenalty(totalZeds),
			TotalPlayersPenalty: calcTotalPlayersPenalty(wave.TotalPlayers),
		}

		res.PredictedDuration = predictDuration(zeds, wave.TotalPlayers, wave.Wave)
		res.PredictedDurationError = wave.DurationRealtime - res.PredictedDuration
		res.KitingPenalty = calcKitingPenalty(wave.DurationRealtime, res.PredictedDuration)

		res.Score = res.ZedsDifficulty * res.WaveSizePenalty * res.TotalPlayersPenalty * res.KitingPenalty

		wave.Result = &res
	}

	{
		res := DifficultyCalculatorGameScore{}

		if len(data.Waves) > 0 {
			res.MinZedsDifficulty = data.Waves[0].Result.ZedsDifficulty
			res.MaxZedsDifficulty = data.Waves[0].Result.ZedsDifficulty
			res.AvgZedsDifficulty = data.Waves[0].Result.ZedsDifficulty

			var (
				zedsDiffSum       = 0.0
				potentialScoreSum = 0.0
				totalWaves        = len(data.Waves)
				hasWaveSkips      = false
				prevWave          = 0
				totalRestarts     = 0
			)

			for _, wave := range data.Waves {
				score := wave.Result

				if wave.Duration > 15 {
					zedsDiffSum += score.ZedsDifficulty
					potentialScoreSum += score.Score

					if score.ZedsDifficulty < res.MinZedsDifficulty {
						res.MinZedsDifficulty = score.ZedsDifficulty
					}
					if score.ZedsDifficulty > res.MaxZedsDifficulty {
						res.MaxZedsDifficulty = score.ZedsDifficulty
					}
					if wave.Attempt > 1 {
						totalRestarts += 1
					}

					if wave.Wave > prevWave+1 {
						hasWaveSkips = true
					}
					prevWave = wave.Wave
				}
			}

			if hasWaveSkips {
				res.CompletionPercent = 0
			} else if session.Status == models.Win {
				res.CompletionPercent = 1
			} else if wave, ok := data.GetLastWave(); ok {
				gameLength := session.Length
				// 25 completed waves for full completion percent for endless mode.
				if session.Mode == models.Endless {
					gameLength = 25
				}
				res.CompletionPercent = npInterp(float64(wave.Wave-1)/float64(gameLength), pair{0, 1}, pair{0.5, 1})
			} else {
				res.CompletionPercent = 0
			}

			res.AvgZedsDifficulty = zedsDiffSum / float64(totalWaves)

			for _, wave := range data.Waves {
				res.StdDevZedsDifficulty += math.Pow(wave.Result.ZedsDifficulty-res.AvgZedsDifficulty, 2)
			}

			res.StdDevZedsDifficulty = res.StdDevZedsDifficulty / float64(totalWaves)
			res.PotentialScore = potentialScoreSum / float64(totalWaves)

			res.RestartsPenalty = npInterp(float64(totalRestarts), pair{0, 3}, pair{1, 0})

			res.FinalScore = res.PotentialScore * res.CompletionPercent * res.RestartsPenalty
		}

		data.Result = &res
	}
}

func (s *DifficultyCalculatorService) getSessions(sessionId []int) ([]*DifficultyCalculatorGame, error) {
	lookup := map[int]*DifficultyCalculatorGame{}

	{
		stmt := fmt.Sprintf(`
			SELECT
				session.id as session_id, 
				session.server_id as server_id,
				session.map_id as map_id,
				session.length as game_length,
				session.diff as game_difficulty,
				session.mode as game_mode,
				session.status as game_status
			FROM session
			WHERE session.id IN (%v)
			`, util.IntArrayToString(sessionId, ","),
		)

		rows, err := s.db.Query(stmt)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			item := DifficultyCalculatorGameSession{}

			err := rows.Scan(
				&item.Id, &item.ServerId, &item.MapId,
				&item.Length, &item.Difficulty, &item.Mode, &item.Status,
			)

			if err != nil {
				return nil, err
			}

			lookup[item.Id] = &DifficultyCalculatorGame{
				Session: &item,
				Waves:   []*DifficultyCalculatorGameWave{},
			}
		}
	}

	{
		stmt := fmt.Sprintf(`
			SELECT DISTINCT
				session.id AS session_id,
				ws.id AS ws_id,

				ws.wave AS wave,
				ws.attempt AS attempt,
				time_to_sec(timediff(ws.completed_at, ws.started_at)) AS duration,

				COUNT(wsp.id) OVER w AS total_players,
				COUNT(CASE WHEN wsp.is_dead THEN wsp.id END) OVER w AS total_deaths,

				SUM(zedtime_length) OVER w AS zedtime_length,
				SUM(zedtime_count) OVER w AS zedtime_count,

				COALESCE(ws_extra.max_monsters, 0) AS max_monsters,
				COALESCE(ws_extra.spawn_cycle, '') AS spawn_cycle,
				LOWER(COALESCE(ws_extra.zeds_type, 'vanilla')) AS zeds_type,

				SUM(k.cyst) OVER w AS cyst, 
				SUM(k.alpha_clot) OVER w AS alpha_clot, 
				SUM(k.slasher) OVER w AS slasher, 
				SUM(k.stalker) OVER w AS stalker, 
				SUM(k.crawler) OVER w AS crawler, 
				SUM(k.gorefast) OVER w AS gorefast, 
				SUM(k.rioter) OVER w AS rioter, 
				SUM(k.elite_crawler) OVER w AS elite_crawler, 
				SUM(k.gorefiend) OVER w AS gorefiend, 
				SUM(k.siren) OVER w AS siren, 
				SUM(k.bloat) OVER w AS bloat, 
				(SUM(k.husk_n) OVER w) + (SUM(k.husk_b) OVER w) AS husk, 
				SUM(k.edar) OVER w AS edar, 
				SUM(k.scrake) OVER w AS scrake, 
				SUM(k.fp) OVER w AS fp,
				SUM(k.qp) OVER w AS qp,
				SUM(k.boss) OVER w AS boss,
				SUM(k.custom) OVER w AS custom
			FROM session
			INNER JOIN wave_stats AS ws
				ON ws.session_id = session.id
			LEFT JOIN wave_stats_extra AS ws_extra
				ON ws_extra.stats_id = ws.id
			INNER JOIN wave_stats_player AS wsp 
				ON wsp.stats_id = ws.id
			INNER JOIN wave_stats_player_kills AS k 
				ON k.player_stats_id = wsp.id
			WHERE
				session.id IN (%v) AND ws.wave <= session.length
			WINDOW w AS (PARTITION BY ws.id)
			`, util.IntArrayToString(sessionId, ","),
		)

		rows, err := s.db.Query(stmt)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			item := DifficultyCalculatorGameWave{
				Zeds: &models.ZedCounter{},
			}

			var sessionId int
			err := rows.Scan(
				&sessionId, &item.Id,
				&item.Wave, &item.Attempt, &item.Duration,
				&item.TotalPlayers, &item.TotalDeaths,
				&item.ZedtimeLength, &item.ZedtimeCount,
				&item.MaxMonsters, &item.SpawnCycle, &item.ZedsType,
				&item.Zeds.Cyst, &item.Zeds.AlphaClot, &item.Zeds.Slasher,
				&item.Zeds.Stalker, &item.Zeds.Crawler, &item.Zeds.Gorefast,
				&item.Zeds.Rioter, &item.Zeds.EliteCrawler, &item.Zeds.Gorefiend,
				&item.Zeds.Siren, &item.Zeds.Bloat, &item.Zeds.Husk, &item.Zeds.Edar,
				&item.Zeds.Scrake, &item.Zeds.FP, &item.Zeds.QP,
				&item.Zeds.Boss, &item.Zeds.Custom,
			)

			if err != nil {
				return nil, err
			}

			item.TotalZeds = item.Zeds.ToMap().GetTotal()

			item.DurationRealtime = float64(item.Duration) - item.ZedtimeLength/5

			if item.TotalZeds > 0 {
				item.MediumPercent = float64(item.Zeds.GetTotalMediums()) / float64(item.TotalZeds)
				item.LargePercent = float64(item.Zeds.GetTotalLarges()) / float64(item.TotalZeds)
			}

			lookup[sessionId].Waves = append(lookup[sessionId].Waves, &item)
		}
	}

	res := []*DifficultyCalculatorGame{}

	for _, value := range lookup {
		res = append(res, value)
	}

	return res, nil
}
