package matches

import (
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/stats"
)

type MatchWave struct {
	WaveId int `json:"wave_id"`

	Wave    int `json:"wave"`
	Attempt int `json:"attempt"`

	Players []*MatchWavePlayer `json:"players"`

	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type MatchWavePlayer struct {
	UserId        int `json:"user_id"`
	PlayerStatsId int `json:"player_stats_id"`

	Perk     models.Perk `json:"perk"`
	Level    int         `json:"level"`
	Prestige int         `json:"prestige"`

	IsDead bool `json:"is_dead"`

	Stats *MatchWavePlayerStats `json:"stats"`
}

type MatchWavePlayerStats struct {
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

	Kills stats.ZedCounter `json:"kills"`

	HuskBackpackKills int `json:"husk_b"`
	HuskRages         int `json:"husk_r"`
}

type MatchSession struct {
	SessionId int `json:"session_id"`
	ServerId  int `json:"server_id"`
	MapId     int `json:"map_id"`

	Mode       models.GameMode       `json:"mode"`
	Length     int                   `json:"length"`
	Difficulty models.GameDifficulty `json:"diff"`

	Status models.GameStatus `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type MatchMap struct {
	Name    *string `json:"name"`
	Preview *string `json:"preview"`
}

type MatchServer struct {
	Name    *string `json:"name"`
	Address *string `json:"address"`
}

type MatchPlayer struct {
	Profile *models.UserProfile `json:"profile"`

	Perk     models.Perk `json:"perk"`
	Level    int         `json:"level"`
	Prestige int         `json:"prestige"`

	Health int `json:"health"`
	Armor  int `json:"armor"`
}

type Match struct {
	Session MatchSession `json:"session"`

	Map    *MatchMap    `json:"map"`
	Server *MatchServer `json:"server"`

	GameData *models.GameData   `json:"game_data"`
	CDData   *models.CDGameData `json:"cd_data"`

	Players    []*MatchPlayer `json:"players"`
	Spectators []*MatchPlayer `json:"spectators"`
}
