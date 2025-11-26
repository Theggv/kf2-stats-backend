package server

import (
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

type RecentUsersResponseUser struct {
	UserProfile *models.UserProfile `json:"user_profile"`

	Match *models.Match `json:"match"`

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
