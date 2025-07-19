package server

import (
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

type AddServerRequest struct {
	Name    string `json:"name" binding:"required"`
	Address string `json:"address" binding:"required"`
}

type AddServerResponse struct {
	Id int `json:"id"`
}

type GetByPatternResponse struct {
	Items []Server `json:"items"`
}

type UpdateNameRequest struct {
	Id   int    `json:"id"`
	Name string `json:"name" binding:"required"`
}

type RecentUsersRequest struct {
	ServerId int `json:"server_id" binding:"required"`

	Pager models.PaginationRequest `json:"pager"`
}

type RecentUsersResponseUserSession struct {
	Id int `json:"id"`

	Mode       models.GameMode       `json:"mode"`
	Length     int                   `json:"length"`
	Difficulty models.GameDifficulty `json:"diff"`

	Status models.GameStatus `json:"status"`

	Wave   int                   `json:"wave"`
	CDData *models.ExtraGameData `json:"cd_data"`

	MapName string `json:"map_name"`

	Perks []int `json:"perks"`

	PlayerId int `json:"-"`
}

type RecentUsersResponseUser struct {
	Id   int    `json:"id"`
	Name string `json:"name"`

	ProfileUrl *string `json:"profile_url"`
	Avatar     *string `json:"avatar"`

	Session *RecentUsersResponseUserSession `json:"session"`

	UpdatedAt *time.Time `json:"updated_at"`

	AuthId string          `json:"-"`
	Type   models.AuthType `json:"-"`

	SessionId         int `json:"-"`
	WaveStatsPlayerId int `json:"-"`
}

type RecentUsersResponse struct {
	Items    []*RecentUsersResponseUser `json:"items"`
	Metadata models.PaginationResponse  `json:"metadata"`
}

type ServerLastSessionResponse struct {
	Id     int               `json:"id"`
	Status models.GameStatus `json:"status"`
}
