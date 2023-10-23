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
			UPDATE session
			SET status = -1
			WHERE id IN (
				SELECT id FROM session
				WHERE status NOT IN ($1, $2) AND id NOT IN (
					SELECT max(id) FROM session
					GROUP BY server_id
				)
			)`, models.Win, models.Lose,
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
			SET status = -1
			WHERE id IN (
				SELECT
					session.id
				FROM session
				INNER JOIN session_game_data gd ON gd.session_id = session.id
				INNER JOIN server ON server.id = session.server_id
				WHERE status IN ($1, $2) and
					(julianday(CURRENT_TIMESTAMP) - julianday(gd.updated_at)) * 24 * 60 > $3
				group by session.id
			)`, models.Lobby, models.InProgress, olderThanMinutes,
		)

		if err != nil {
			fmt.Printf("[abortOldMatches] error: %v\n", err)
		}
	}
}
