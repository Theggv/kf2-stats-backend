package cron

import (
	"github.com/theggv/kf2-stats-backend/pkg/common/store"
)

func SetupTasks(s *store.Store) {
	go detectDroppedSessions(s.Db)
	go abortOldMatches(s.Db)
	go deleteEmptySessions(s.Db)

	go setupProcessDemosTask(s)
}
