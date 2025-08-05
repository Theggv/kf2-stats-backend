package cron

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/demorecord"
	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/store"
	"github.com/theggv/kf2-stats-backend/pkg/session"
)

func setupProcessDemosTask(s *store.Store) {
	for range time.Tick(15 * time.Second) {
		err := processDemos(s)
		if err != nil {
			fmt.Printf("[processDemos] error: %v\n", err)
		}
	}
}

func getDemoCount(db *sql.DB) (int, error) {
	var count int
	{
		row := db.QueryRow(`SELECT COUNT(*) FROM session_demo WHERE processed = 0`)
		err := row.Scan(&count)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func getDemo(sessionId int, s *session.SessionService) (*demorecord.DemoRecordParsed, error) {
	rawDemo, err := s.GetDemo(sessionId)
	if err != nil {
		return nil, err
	}

	parsedDemo, err := rawDemo.ToParsed()
	if err != nil {
		return nil, err
	}

	s.LoadDemoUsers(parsedDemo)

	return parsedDemo, nil
}

func processDemo(sessionId int, analysis *demorecord.DemoRecordAnalysis, db *sql.DB) error {
	_, err := db.Exec(`
		UPDATE session_demo SET processed = 1 WHERE session_id = ?`, sessionId,
	)
	if err != nil {
		return err
	}

	if analysis.Analytics.BuffsUptime.BuffedTicks > 0 {
		type buffs struct {
			buffed, total int
		}

		buffsData := map[int]*buffs{}

		for _, wave := range analysis.Waves {
			if wave.Analytics.BuffsUptime.BuffedTicks <= 0 {
				continue
			}

			for _, player := range wave.PlayerEvents.Perks {
				if player.Perk != models.Medic {
					continue
				}

				if data, ok := buffsData[player.UserId]; ok {
					data.buffed += wave.Analytics.BuffsUptime.BuffedTicks
					data.total += wave.Analytics.BuffsUptime.TotalTicks
				} else {
					buffsData[player.UserId] = &buffs{
						buffed: wave.Analytics.BuffsUptime.BuffedTicks,
						total:  wave.Analytics.BuffsUptime.TotalTicks,
					}
				}
			}
		}

		for userIndex, item := range buffsData {
			if profile := analysis.Players.GetByIndex(userIndex); profile != nil {
				_, err = db.Exec(`
							UPDATE session_aggregated
							SET buffs_active_length = ?, buffs_total_length = ?
							WHERE session_id = ? AND user_id = ? AND perk = 3`,
					float64(item.buffed)/100.0, float64(item.total)/100,
					sessionId, profile.Id,
				)

				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func processDemos(s *store.Store) error {
	count, err := getDemoCount(s.Db)
	if err != nil {
		return err
	}

	if count > 0 {
		fmt.Printf("[processDemos] found %v demos\n", count)

	}

	for count > 0 {
		rows, err := s.Db.Query(`
			SELECT session_id 
			FROM session_demo WHERE processed = 0 LIMIT 100`,
		)

		if err != nil {
			return err
		}

		for rows.Next() {
			var sessionId int
			err = rows.Scan(&sessionId)
			if err != nil {
				return err
			}

			demo, err := getDemo(sessionId, s.Sessions)
			if err != nil {
				return err
			}

			analysis := demo.Analyze()

			count -= 1

			err = processDemo(sessionId, analysis, s.Db)
			if err != nil {
				return err
			}

			if count > 0 && count%100 == 0 {
				fmt.Printf("[processDemos] demos %v left to process\n", count)
			}
		}
	}

	return nil
}
