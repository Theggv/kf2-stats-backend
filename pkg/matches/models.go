package matches

import (
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

type Match struct {
	Session MatchSession `json:"session"`

	Details MatchDetails `json:"details"`

	Metadata models.SessionMetadata `json:"metadata"`
}

type MatchSession struct {
	Id int `json:"id"`

	ServerId int `json:"server_id"`
	MapId    int `json:"map_id"`

	Mode       models.GameMode       `json:"mode"`
	Length     int                   `json:"length"`
	Difficulty models.GameDifficulty `json:"diff"`

	Status models.GameStatus `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type MatchDetails struct {
	Map    *MatchMap    `json:"map,omitempty"`
	Server *MatchServer `json:"server,omitempty"`

	GameData      *models.GameData      `json:"game_data,omitempty"`
	ExtraGameData *models.ExtraGameData `json:"extra_data,omitempty"`

	LiveData *MatchLiveData `json:"live_data,omitempty"`

	UserData *MatchUserData `json:"user_data,omitempty"`
}

type MatchMap struct {
	Name    string  `json:"name"`
	Preview *string `json:"preview"`
}

type MatchServer struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type MatchLiveData struct {
	Players    []*MatchPlayer `json:"players"`
	Spectators []*MatchPlayer `json:"spectators"`
}

type MatchPlayer struct {
	Profile *models.UserProfile `json:"profile"`

	Perk     models.Perk `json:"perk"`
	Level    int         `json:"level"`
	Prestige int         `json:"prestige"`

	Health int `json:"health"`
	Armor  int `json:"armor"`
}

type MatchUserData struct {
	LastSeen *time.Time `json:"last_seen"`

	Perks []int `json:"perks"`

	Stats MatchUserDataStats `json:"stats"`
}

type MatchUserDataStats struct {
	DamageDealt int `json:"damage_dealt"`
}

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

	Kills models.ZedCounter `json:"kills"`

	HuskBackpackKills int `json:"husk_b"`
	HuskRages         int `json:"husk_r"`
}
