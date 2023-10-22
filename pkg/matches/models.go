package matches

import (
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/session"
	"github.com/theggv/kf2-stats-backend/pkg/users"
)

type Player struct {
	Id       int            `json:"id"`
	AuthId   string         `json:"auth_id"`
	AuthType users.AuthType `json:"auth_type"`
	Name     string         `json:"name"`

	PlayerStatsId int `json:"player_stats_id"`

	Perk     models.Perk `json:"perk"`
	Level    int         `json:"level"`
	Prestige int         `json:"prestige"`

	IsDead bool `json:"is_dead"`
}

type PlayerWithSteamData struct {
	Id int `json:"id"`

	Name       string  `json:"name"`
	ProfileUrl *string `json:"profile_url"`
	Avatar     *string `json:"avatar"`

	PlayerStatsId int `json:"player_stats_id"`

	Perk     models.Perk `json:"perk"`
	Level    int         `json:"level"`
	Prestige int         `json:"prestige"`

	IsDead bool `json:"is_dead"`
}

type MatchWave struct {
	Id      int `json:"id"`
	Wave    int `json:"wave"`
	Attempt int `json:"attempt"`

	Players []PlayerWithSteamData `json:"players"`

	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
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

type Match struct {
	Session  MatchSession       `json:"session"`
	Map      *MatchMap          `json:"map"`
	Server   *MatchServer       `json:"server"`
	GameData *session.GameData  `json:"game_data"`
	CDData   *models.CDGameData `json:"cd_data"`
}
