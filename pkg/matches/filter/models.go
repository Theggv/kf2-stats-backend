package filter

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
	Map    *MatchMap    `json:"map"`
	Server *MatchServer `json:"server"`

	GameData      *models.GameData      `json:"game_data"`
	ExtraGameData *models.ExtraGameData `json:"extra_data"`

	LiveData *MatchLiveData `json:"live_data"`

	UserData *MatchUserData `json:"user_data"`
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
	Perks []int `json:"perks"`

	Stats MatchUserDataStats `json:"stats"`
}

type MatchUserDataStats struct {
	DamageDealt int `json:"damage_dealt"`
}
