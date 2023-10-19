package session

import (
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

type SessionMap struct {
	Name    *string `json:"name"`
	Preview *string `json:"preview"`
}

type SessionServer struct {
	Name    *string `json:"name"`
	Address *string `json:"address"`
}

type SessionGameDetails struct {
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

	Map    *SessionMap    `json:"map"`
	Server *SessionServer `json:"server"`

	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}
