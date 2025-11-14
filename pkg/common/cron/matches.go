package cron

import (
	"database/sql"
	"fmt"
	"time"
)

func handleDanglingSessions(db *sql.DB) {
	olderThanMinutes := 15

	for range time.Tick(1 * time.Minute) {
		_, err := db.Exec(`CALL handle_dangling_sessions(?)`, olderThanMinutes)

		if err != nil {
			fmt.Printf("[abortOldMatches] error: %v\n", err)
		}
	}
}
