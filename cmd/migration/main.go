package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/config"
	"github.com/theggv/kf2-stats-backend/pkg/common/database"
	"github.com/theggv/kf2-stats-backend/pkg/common/database/mysql"
	"github.com/theggv/kf2-stats-backend/pkg/maps"
	"github.com/theggv/kf2-stats-backend/pkg/server"
	"github.com/theggv/kf2-stats-backend/pkg/session"
	"github.com/theggv/kf2-stats-backend/pkg/users"
)

func executeInsert[T interface{}](
	mysql *sql.DB,
	baseQuery string,
	valuePattern string,
	items []T,
	prepareArgsFunc func(T) []interface{},
) {
	values := make([]string, 0)
	args := make([]interface{}, 0)

	for _, item := range items {
		values = append(values, valuePattern)
		args = append(args, prepareArgsFunc(item)...)
	}

	sql := fmt.Sprintf("%v VALUES\n%v",
		baseQuery,
		strings.Join(values, ","),
	)

	_, err := mysql.Exec(sql, args...)
	if err != nil {
		panic(err)
	}
}

func insertServers(sqlite, mysql *sql.DB) {
	fmt.Print("Inserting servers...\n")

	rows, err := sqlite.Query("SELECT id, name, address FROM server")
	if err != nil {
		panic(err)
	}

	chunkSize := 100
	count := 0

	items := make([]server.Server, 0)

	insert := func() {
		executeInsert(mysql,
			"INSERT IGNORE INTO server (id, name, address)",
			"(?, ?, ?)",
			items,
			func(i server.Server) (out []interface{}) {
				out = append(out, i.Id, i.Name, i.Address)
				return
			},
		)
	}

	for rows.Next() {
		if count == chunkSize {
			insert()
			items = items[:0]
			count = 0
		}

		item := server.Server{}
		rows.Scan(&item.Id, &item.Name, &item.Address)

		count += 1
		items = append(items, item)
	}

	if len(items) > 0 {
		insert()
	}
}

func insertMaps(sqlite, mysql *sql.DB) {
	fmt.Print("Inserting maps...\n")

	rows, err := sqlite.Query("SELECT id, name, preview FROM maps")
	if err != nil {
		panic(err)
	}

	chunkSize := 100
	count := 0

	items := make([]maps.Map, 0)

	insert := func() {
		executeInsert(mysql,
			"INSERT IGNORE INTO maps (id, name, preview)",
			"(?, ?, ?)",
			items,
			func(i maps.Map) (out []interface{}) {
				out = append(out, i.Id, i.Name, i.Preview)
				return
			},
		)
	}

	for rows.Next() {
		if count == chunkSize {
			insert()
			items = items[:0]
			count = 0
		}

		item := maps.Map{}
		rows.Scan(&item.Id, &item.Name, &item.Preview)

		count += 1
		items = append(items, item)
	}

	if len(items) > 0 {
		insert()
	}
}

func insertUsers(sqlite, mysql *sql.DB) {
	fmt.Print("Inserting users...\n")

	rows, err := sqlite.Query("SELECT id, auth_id, auth_type, name FROM users")
	if err != nil {
		panic(err)
	}

	chunkSize := 100
	count := 0

	items := make([]users.User, 0)

	insert := func() {
		executeInsert(mysql,
			"INSERT IGNORE INTO users (id, auth_id, auth_type, name)",
			"(?, ?, ?, ?)",
			items,
			func(i users.User) (out []interface{}) {
				out = append(out, i.Id, i.AuthId, i.Type, i.Name)
				return
			},
		)
	}

	for rows.Next() {
		if count == chunkSize {
			insert()
			items = items[:0]
			count = 0
		}

		item := users.User{}
		rows.Scan(&item.Id, &item.AuthId, &item.Type, &item.Name)

		count += 1
		items = append(items, item)
	}

	if len(items) > 0 {
		insert()
	}
}

func insertSessions(sqlite, mysql *sql.DB) {
	fmt.Print("Inserting sessions...\n")

	rows, err := sqlite.Query(`
		SELECT 
			id, server_id, map_id, 
			mode, length, diff, status, 
			created_at, updated_at, started_at, completed_at
		FROM session`,
	)
	if err != nil {
		panic(err)
	}

	chunkSize := 100
	count := 0

	items := make([]session.Session, 0)

	insert := func() {
		executeInsert(mysql,
			`INSERT IGNORE INTO session (id, server_id, map_id, 
				mode, length, diff, status, 
				created_at, updated_at, started_at, completed_at)`,
			"(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			items,
			func(i session.Session) (out []interface{}) {
				out = append(out,
					i.Id, i.ServerId, i.MapId,
					i.Mode, i.Length, i.Difficulty, i.Status,
					i.CreatedAt, i.UpdatedAt, i.StartedAt, i.CompletedAt,
				)
				return
			},
		)
	}

	for rows.Next() {
		if count == chunkSize {
			insert()
			items = items[:0]
			count = 0
		}

		item := session.Session{}
		rows.Scan(
			&item.Id, &item.ServerId, &item.MapId,
			&item.Mode, &item.Length, &item.Difficulty, &item.Status,
			&item.CreatedAt, &item.UpdatedAt, &item.StartedAt, &item.CompletedAt,
		)

		count += 1
		items = append(items, item)
	}

	if len(items) > 0 {
		insert()
	}
}

func insertSessionGameData(sqlite, mysql *sql.DB) {
	fmt.Print("Inserting session game data...\n")

	type data struct {
		SessionId     int        `json:"session_id"`
		MaxPlayers    int        `json:"max_players"`
		PlayersOnline int        `json:"players_online"`
		PlayersAlive  int        `json:"players_alive"`
		Wave          int        `json:"wave"`
		IsTraderTime  int        `json:"is_trader_time"`
		ZedsLeft      int        `json:"zeds_left"`
		UpdatedAt     *time.Time `json:"updated_at"`
	}

	rows, err := sqlite.Query(`
		SELECT 
			session_id, max_players, players_online, players_alive, 
			wave, is_trader_time, zeds_left, updated_at
		FROM session_game_data`,
	)
	if err != nil {
		panic(err)
	}

	chunkSize := 100
	count := 0

	items := make([]data, 0)

	insert := func() {
		executeInsert(mysql,
			`INSERT IGNORE INTO session_game_data (session_id, max_players, players_online, players_alive, 
				wave, is_trader_time, zeds_left, updated_at)`,
			"(?, ?, ?, ?, ?, ?, ?, ?)",
			items,
			func(i data) (out []interface{}) {
				out = append(out,
					i.SessionId, i.MaxPlayers, i.PlayersOnline, i.PlayersAlive,
					i.Wave, i.IsTraderTime, i.ZedsLeft, i.UpdatedAt,
				)
				return
			},
		)
	}

	for rows.Next() {
		if count == chunkSize {
			insert()
			items = items[:0]
			count = 0
		}

		item := data{}
		rows.Scan(
			&item.SessionId, &item.MaxPlayers, &item.PlayersOnline, &item.PlayersAlive,
			&item.Wave, &item.IsTraderTime, &item.ZedsLeft, &item.UpdatedAt,
		)

		count += 1
		items = append(items, item)
	}

	if len(items) > 0 {
		insert()
	}
}

func insertSessionGameDataCD(sqlite, mysql *sql.DB) {
	fmt.Print("Inserting session cd data...\n")

	type data struct {
		SessionId     int        `json:"session_id"`
		SpawnCycle    string     `json:"spawn_cycle"`
		MaxMonsters   int        `json:"max_monsters"`
		WaveSizeFakes int        `json:"wave_size_fakes"`
		ZedsType      string     `json:"zeds_type"`
		UpdatedAt     *time.Time `json:"updated_at"`
	}

	rows, err := sqlite.Query(`
		SELECT 
			session_id, spawn_cycle, max_monsters, 
			wave_size_fakes, zeds_type, updated_at
		FROM session_game_data_cd`,
	)
	if err != nil {
		panic(err)
	}

	chunkSize := 100
	count := 0

	items := make([]data, 0)

	insert := func() {
		executeInsert(mysql,
			`INSERT IGNORE INTO session_game_data_cd (session_id, spawn_cycle, max_monsters, 
				wave_size_fakes, zeds_type, updated_at)`,
			"(?, ?, ?, ?, ?, ?)",
			items,
			func(i data) (out []interface{}) {
				out = append(out,
					i.SessionId, i.SpawnCycle, i.MaxMonsters,
					i.WaveSizeFakes, i.ZedsType, i.UpdatedAt,
				)
				return
			},
		)
	}

	for rows.Next() {
		if count == chunkSize {
			insert()
			items = items[:0]
			count = 0
		}

		item := data{}
		rows.Scan(
			&item.SessionId, &item.SpawnCycle, &item.MaxMonsters,
			&item.WaveSizeFakes, &item.ZedsType, &item.UpdatedAt,
		)

		count += 1
		items = append(items, item)
	}

	if len(items) > 0 {
		insert()
	}
}

func insertWaveStats(sqlite, mysql *sql.DB) {
	fmt.Print("Inserting wave stats...\n")

	type data struct {
		Id          int        `json:"id"`
		SessionId   int        `json:"session_id"`
		Wave        int        `json:"wave"`
		Attempt     int        `json:"attempt"`
		StartedAt   *time.Time `json:"started_at"`
		CompletedAt *time.Time `json:"completed_at"`
	}

	rows, err := sqlite.Query(`
		SELECT 
			id, session_id, wave, 
			attempt, started_at, completed_at
		FROM wave_stats`,
	)
	if err != nil {
		panic(err)
	}

	chunkSize := 100
	count := 0

	items := make([]data, 0)

	insert := func() {
		executeInsert(mysql,
			`INSERT IGNORE INTO wave_stats (id, session_id, wave, 
				attempt, started_at, completed_at)`,
			"(?, ?, ?, ?, ?, ?)",
			items,
			func(i data) (out []interface{}) {
				out = append(out,
					i.Id, i.SessionId, i.Wave,
					i.Attempt, i.StartedAt, i.CompletedAt,
				)
				return
			},
		)
	}

	for rows.Next() {
		if count == chunkSize {
			insert()
			items = items[:0]
			count = 0
		}

		item := data{}
		rows.Scan(
			&item.Id, &item.SessionId, &item.Wave,
			&item.Attempt, &item.StartedAt, &item.CompletedAt,
		)

		count += 1
		items = append(items, item)
	}

	if len(items) > 0 {
		insert()
	}
}

func insertWaveStatsCD(sqlite, mysql *sql.DB) {
	fmt.Print("Inserting cd wave stats...\n")

	type data struct {
		StatsId       int        `json:"stats_id"`
		SpawnCycle    string     `json:"spawn_cycle"`
		MaxMonsters   int        `json:"max_monsters"`
		WaveSizeFakes int        `json:"wave_size_fakes"`
		ZedsType      string     `json:"zeds_type"`
		CreatedAt     *time.Time `json:"created_at"`
	}

	rows, err := sqlite.Query(`
		SELECT 
			stats_id, spawn_cycle, max_monsters, 
			wave_size_fakes, zeds_type, created_at
		FROM wave_stats_cd`,
	)
	if err != nil {
		panic(err)
	}

	chunkSize := 100
	count := 0

	items := make([]data, 0)

	insert := func() {
		executeInsert(mysql,
			`INSERT IGNORE INTO wave_stats_cd (stats_id, spawn_cycle, max_monsters, 
				wave_size_fakes, zeds_type, created_at)`,
			"(?, ?, ?, ?, ?, ?)",
			items,
			func(i data) (out []interface{}) {
				out = append(out,
					i.StatsId, i.SpawnCycle, i.MaxMonsters,
					i.WaveSizeFakes, i.ZedsType, i.CreatedAt,
				)
				return
			},
		)
	}

	for rows.Next() {
		if count == chunkSize {
			insert()
			items = items[:0]
			count = 0
		}

		item := data{}
		rows.Scan(
			&item.StatsId, &item.SpawnCycle, &item.MaxMonsters,
			&item.WaveSizeFakes, &item.ZedsType, &item.CreatedAt,
		)

		count += 1
		items = append(items, item)
	}

	if len(items) > 0 {
		insert()
	}
}

func insertWaveStatsPlayer(sqlite, mysql *sql.DB) {
	fmt.Print("Inserting player wave stats...\n")

	type data struct {
		Id       int `json:"id"`
		StatsId  int `json:"stats_id"`
		PlayerId int `json:"player_id"`

		Perk     int `json:"perk"`
		Level    int `json:"level"`
		Prestige int `json:"prestige"`

		IsDead bool `json:"is_dead"`

		ShotsFired int `json:"shots_fired"`
		ShotsHit   int `json:"shots_hit"`
		ShotsHS    int `json:"shots_hs"`

		DoshEarned int `json:"dosh_earned"`

		HealsGiven    int `json:"heals_given"`
		HealsReceived int `json:"heals_recv"`

		DamageDealt int `json:"damage_dealt"`
		DamageTaken int `json:"damage_taken"`

		ZedTimeCount  int     `json:"zedtime_count"`
		ZedTimeLength float32 `json:"zedtime_length"`

		CreatedAt *time.Time `json:"created_at"`
	}

	rows, err := sqlite.Query(`
		SELECT 
			id, stats_id, player_id, 
			perk, level, prestige, is_dead, 
			shots_fired, shots_hit, shots_hs, 
			dosh_earned, heals_given, heals_recv, 
			damage_dealt, damage_taken, 
			zedtime_count, zedtime_length, created_at
		FROM wave_stats_player`,
	)
	if err != nil {
		panic(err)
	}

	chunkSize := 100
	count := 0

	items := make([]data, 0)

	insert := func() {
		executeInsert(mysql,
			`INSERT IGNORE INTO wave_stats_player (id, stats_id, player_id, 
				perk, level, prestige, is_dead, 
				shots_fired, shots_hit, shots_hs, 
				dosh_earned, heals_given, heals_recv, 
				damage_dealt, damage_taken, 
				zedtime_count, zedtime_length, created_at)`,
			"(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			items,
			func(i data) (out []interface{}) {
				out = append(out,
					i.Id, i.StatsId, i.PlayerId, i.Perk,
					i.Level, i.Prestige, i.IsDead,
					i.ShotsFired, i.ShotsHit, i.ShotsHS,
					i.DoshEarned, i.HealsGiven, i.HealsReceived,
					i.DamageDealt, i.DamageTaken,
					i.ZedTimeCount, i.ZedTimeLength, i.CreatedAt,
				)
				return
			},
		)
	}

	for rows.Next() {
		if count == chunkSize {
			insert()
			items = items[:0]
			count = 0
		}

		item := data{}
		rows.Scan(
			&item.Id, &item.StatsId, &item.PlayerId, &item.Perk,
			&item.Level, &item.Prestige, &item.IsDead,
			&item.ShotsFired, &item.ShotsHit, &item.ShotsHS,
			&item.DoshEarned, &item.HealsGiven, &item.HealsReceived,
			&item.DamageDealt, &item.DamageTaken,
			&item.ZedTimeCount, &item.ZedTimeLength, &item.CreatedAt,
		)

		count += 1
		items = append(items, item)
	}

	if len(items) > 0 {
		insert()
	}
}

func insertWaveStatsPlayerKills(sqlite, mysql *sql.DB) {
	fmt.Print("Inserting player kills...\n")

	type data struct {
		PlayerStatsId int `json:"player_stats_id"`
		Cyst          int `json:"cyst"`
		AlphaClot     int `json:"alpha_clot"`
		Slasher       int `json:"slasher"`
		Stalker       int `json:"stalker"`
		Crawler       int `json:"crawler"`
		Gorefast      int `json:"gorefast"`
		Rioter        int `json:"rioter"`
		EliteCrawler  int `json:"elite_crawler"`
		Gorefiend     int `json:"gorefiend"`
		Siren         int `json:"siren"`
		Bloat         int `json:"bloat"`
		Edar          int `json:"edar"`
		HuskN         int `json:"husk_n"`
		HuskB         int `json:"husk_b"`
		HuskR         int `json:"husk_r"`
		Scrake        int `json:"scrake"`
		Fp            int `json:"fp"`
		Qp            int `json:"qp"`
		Boss          int `json:"boss"`
		Custom        int `json:"custom"`
	}

	rows, err := sqlite.Query(`
		SELECT 
			player_stats_id, 
			cyst, alpha_clot, slasher, stalker, crawler, 
			gorefast, rioter, elite_crawler, gorefiend, 
			siren, bloat, edar, husk_n, husk_b, husk_r, 
			scrake, fp, qp, boss
		FROM wave_stats_player_kills`,
	)
	if err != nil {
		panic(err)
	}

	chunkSize := 100
	count := 0

	items := make([]data, 0)

	insert := func() {
		executeInsert(mysql,
			`INSERT IGNORE INTO wave_stats_player_kills (player_stats_id, 
				cyst, alpha_clot, slasher, stalker, crawler, 
				gorefast, rioter, elite_crawler, gorefiend, 
				siren, bloat, edar, husk_n, husk_b, husk_r, 
				scrake, fp, qp, boss, custom)`,
			"(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)",
			items,
			func(i data) (out []interface{}) {
				out = append(out,
					i.PlayerStatsId,
					i.Cyst, i.AlphaClot, i.Slasher, i.Stalker, i.Crawler,
					i.Gorefast, i.Rioter, i.EliteCrawler, i.Gorefiend,
					i.Siren, i.Bloat, i.Edar, i.HuskN, i.HuskB, i.HuskR,
					i.Scrake, i.Fp, i.Qp, i.Boss,
				)
				return
			},
		)
	}

	for rows.Next() {
		if count == chunkSize {
			insert()
			items = items[:0]
			count = 0
		}

		item := data{}
		rows.Scan(
			&item.PlayerStatsId,
			&item.Cyst, &item.AlphaClot, &item.Slasher, &item.Stalker, &item.Crawler,
			&item.Gorefast, &item.Rioter, &item.EliteCrawler, &item.Gorefiend,
			&item.Siren, &item.Bloat, &item.Edar, &item.HuskN, &item.HuskB, &item.HuskR,
			&item.Scrake, &item.Fp, &item.Qp, &item.Boss,
		)

		count += 1
		items = append(items, item)
	}

	if len(items) > 0 {
		insert()
	}
}

func insertWaveStatsPlayerInjuredBy(sqlite, mysql *sql.DB) {
	fmt.Print("Inserting player injured by...\n")

	type data struct {
		PlayerStatsId int `json:"player_stats_id"`
		Cyst          int `json:"cyst"`
		AlphaClot     int `json:"alpha_clot"`
		Slasher       int `json:"slasher"`
		Stalker       int `json:"stalker"`
		Crawler       int `json:"crawler"`
		Gorefast      int `json:"gorefast"`
		Rioter        int `json:"rioter"`
		EliteCrawler  int `json:"elite_crawler"`
		Gorefiend     int `json:"gorefiend"`
		Siren         int `json:"siren"`
		Bloat         int `json:"bloat"`
		Edar          int `json:"edar"`
		Husk          int `json:"husk"`
		Scrake        int `json:"scrake"`
		Fp            int `json:"fp"`
		Qp            int `json:"qp"`
		Boss          int `json:"boss"`
	}

	rows, err := sqlite.Query(`
		SELECT 
			player_stats_id, 
			cyst, alpha_clot, slasher, stalker, crawler, 
			gorefast, rioter, elite_crawler, gorefiend, 
			siren, bloat, edar, husk, 
			scrake, fp, qp, boss
		FROM wave_stats_player_injured_by`,
	)
	if err != nil {
		panic(err)
	}

	chunkSize := 100
	count := 0

	items := make([]data, 0)

	insert := func() {
		executeInsert(mysql,
			`INSERT IGNORE INTO wave_stats_player_injured_by (player_stats_id, 
				cyst, alpha_clot, slasher, stalker, crawler, 
				gorefast, rioter, elite_crawler, gorefiend, 
				siren, bloat, edar, husk,
				scrake, fp, qp, boss)`,
			"(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			items,
			func(i data) (out []interface{}) {
				out = append(out,
					i.PlayerStatsId,
					i.Cyst, i.AlphaClot, i.Slasher, i.Stalker, i.Crawler,
					i.Gorefast, i.Rioter, i.EliteCrawler, i.Gorefiend,
					i.Siren, i.Bloat, i.Edar, i.Husk,
					i.Scrake, i.Fp, i.Qp, i.Boss,
				)
				return
			},
		)
	}

	for rows.Next() {
		if count == chunkSize {
			insert()
			items = items[:0]
			count = 0
		}

		item := data{}
		rows.Scan(
			&item.PlayerStatsId,
			&item.Cyst, &item.AlphaClot, &item.Slasher, &item.Stalker, &item.Crawler,
			&item.Gorefast, &item.Rioter, &item.EliteCrawler, &item.Gorefiend,
			&item.Siren, &item.Bloat, &item.Edar, &item.Husk,
			&item.Scrake, &item.Fp, &item.Qp, &item.Boss,
		)

		count += 1
		items = append(items, item)
	}

	if len(items) > 0 {
		insert()
	}
}

func insertUsersActivity(sqlite, mysql *sql.DB) {
	fmt.Print("Inserting users activity...\n")

	_, err := mysql.Exec(`
		INSERT INTO users_activity (
			user_id, current_session_id, last_session_id)
		SELECT id, NULL, NULL
		FROM users`,
	)

	if err != nil {
		panic(err)
	}
}

func insertWaveStatsPlayerComms(sqlite, mysql *sql.DB) {
	fmt.Print("Inserting player comms...\n")

	_, err := mysql.Exec(`
		INSERT INTO wave_stats_player_comms (player_stats_id,
			request_healing, request_dosh, request_help, 
			taunt_zeds, follow_me, get_to_the_trader, 
			affirmative, negative, thank_you)
		SELECT id, 0, 0, 0, 0, 0, 0, 0, 0, 0 
		FROM wave_stats_player`,
	)
	if err != nil {
		panic(err)
	}
}

func main() {
	defer func(begin time.Time) {
		fmt.Printf("Done. %vs\n", time.Since(begin).Seconds())
	}(time.Now())

	config := config.Instance
	mysql := mysql.NewDBInstance(
		config.DBUser, config.DBPassword, config.DBHost, config.DBName, config.DBPort,
	)

	sqlite := database.NewSQLiteDB(config.DatabasePath)

	insertServers(sqlite, mysql)
	insertMaps(sqlite, mysql)
	insertUsers(sqlite, mysql)
	insertSessions(sqlite, mysql)
	insertSessionGameData(sqlite, mysql)
	insertSessionGameDataCD(sqlite, mysql)
	insertWaveStats(sqlite, mysql)
	insertWaveStatsCD(sqlite, mysql)
	insertWaveStatsPlayer(sqlite, mysql)
	insertWaveStatsPlayerKills(sqlite, mysql)
	insertWaveStatsPlayerInjuredBy(sqlite, mysql)
	insertUsersActivity(sqlite, mysql)
	insertWaveStatsPlayerComms(sqlite, mysql)
}
