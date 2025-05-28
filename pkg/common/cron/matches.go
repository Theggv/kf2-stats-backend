package cron

import (
	"database/sql"
	"fmt"
	"time"
)

func detectDroppedSessions(db *sql.DB) {
	for range time.Tick(3 * time.Minute) {
		_, err := db.Exec(`CALL fix_dropped_sessions()`)

		if err != nil {
			fmt.Printf("[detectDroppedSessions] error: %v\n", err)
		}
	}
}

func abortOldMatches(db *sql.DB) {
	olderThanMinutes := 15

	for range time.Tick(3 * time.Minute) {
		_, err := db.Exec(`CALL abort_old_sessions(?)`, olderThanMinutes)

		if err != nil {
			fmt.Printf("[abortOldMatches] error: %v\n", err)
		}
	}
}

func deleteEmptySessions(db *sql.DB) {
	for range time.Tick(60 * time.Minute) {
		_, err := db.Exec(`CALL delete_empty_sessions()`)

		if err != nil {
			fmt.Printf("[deleteEmptySessions] error: %v\n", err)
		}
	}
}
