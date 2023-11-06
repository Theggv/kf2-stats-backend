package matches

import (
	"fmt"
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

func (s *MatchesService) setupTasks() {
	go detectDroppedSessions(s)
	go abortOldMatches(s)
}

func detectDroppedSessions(s *MatchesService) {
	for range time.Tick(3 * time.Minute) {
		_, err := s.db.Exec(`
			UPDATE session, (
				SELECT max(id) as max_id FROM session
				GROUP BY server_id
			) as tbl
			SET status = -1
			WHERE 
				session.id <> 0 AND session.id NOT IN (tbl.max_id) AND 
				status NOT IN (?, ?)`,
			models.Win, models.Lose,
		)

		if err != nil {
			fmt.Printf("[detectDroppedSessions] error: %v\n", err)
		}
	}
}

func abortOldMatches(s *MatchesService) {
	olderThanMinutes := 15

	for range time.Tick(3 * time.Minute) {
		_, err := s.db.Exec(`
			UPDATE session
			INNER JOIN server ON server.id = session.server_id
			INNER JOIN session_game_data gd ON gd.session_id = session.id
			SET session.status = -1
			WHERE 
				session.id <> 0 AND 
				session.status IN (0, 1) AND 
				timestampdiff(MINUTE, gd.updated_at, CURRENT_TIMESTAMP) > ?`,
			olderThanMinutes,
		)

		if err != nil {
			fmt.Printf("[abortOldMatches] error: %v\n", err)
		}
	}
}
