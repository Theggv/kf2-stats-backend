package session

import (
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

type GameData struct {
	MaxPlayers    int `json:"max_players"`
	PlayersOnline int `json:"players_online"`
	PlayersAlive  int `json:"players_alive"`

	Wave         int  `json:"wave"`
	IsTraderTime bool `json:"is_trader_time"`
	ZedsLeft     int  `json:"zeds_left"`
}

type Session struct {
	Id       int `json:"id"`
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
