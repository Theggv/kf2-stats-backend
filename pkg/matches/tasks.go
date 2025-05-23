package matches

import (
	"fmt"
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/demorecord"
	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/session"
)

func (s *MatchesService) setupTasks() {
	go detectDroppedSessions(s)
	go abortOldMatches(s)
	go deleteEmptySessions(s)
	go setupProcessDemosTask(s)
}

func detectDroppedSessions(s *MatchesService) {
	for range time.Tick(3 * time.Minute) {
		_, err := s.db.Exec(`CALL fix_dropped_sessions()`)

		if err != nil {
			fmt.Printf("[detectDroppedSessions] error: %v\n", err)
		}
	}
}

func abortOldMatches(s *MatchesService) {
	olderThanMinutes := 15

	for range time.Tick(3 * time.Minute) {
		_, err := s.db.Exec(`CALL abort_old_sessions(?)`, olderThanMinutes)

		if err != nil {
			fmt.Printf("[abortOldMatches] error: %v\n", err)
		}
	}
}

func deleteEmptySessions(s *MatchesService) {
	for range time.Tick(60 * time.Minute) {
		_, err := s.db.Exec(`CALL delete_empty_sessions()`)

		if err != nil {
			fmt.Printf("[deleteEmptySessions] error: %v\n", err)
		}
	}
}

func setupProcessDemosTask(s *MatchesService) {
	for range time.Tick(10 * time.Second) {
		err := processDemos(s)
		if err != nil {
			fmt.Printf("[processDemos] error: %v\n", err)
		}
	}
}

func processDemos(s *MatchesService) error {
	var count int
	{
		row := s.db.QueryRow(`SELECT COUNT(*) FROM session_demo WHERE processed = 0`)
		err := row.Scan(&count)
		if err != nil {
			return err
		}
	}

	for count > 0 {
		rows, err := s.db.Query(`
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

			demo, err := getDemo(sessionId, s.sessionService)
			if err != nil {
				return err
			}

			analysis := demo.Analyze()

			count -= 1

			_, err = s.db.Exec(`
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
					fmt.Printf("user_idx=%v data=%v\n", userIndex, item)
					if profile := analysis.Players.GetByIndex(userIndex); profile != nil {
						_, err = s.db.Exec(`
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
		}
	}

	return nil
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
