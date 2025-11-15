package cron

import (
	"github.com/theggv/kf2-stats-backend/pkg/common/store"
)

func SetupTasks(s *store.Store) {
	go handleDanglingSessions(s.Db)

	go setupProcessDemosTask(s)
}
