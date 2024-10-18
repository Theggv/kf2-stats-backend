package matches

import (
	"fmt"
	"time"
)

func (s *MatchesService) setupTasks() {
	go detectDroppedSessions(s)
	go abortOldMatches(s)
	go deleteEmptySessions(s)
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
