package server

import (
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/analytics"
	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

type SessionCountRequest struct {
	ServerId int `json:"server_id"`

	From   *time.Time           `json:"date_from"`
	To     *time.Time           `json:"date_to"`
	Period analytics.TimePeriod `json:"period" binding:"required"`
}

type SessionCountResponse struct {
	Items []*models.PeriodData `json:"items"`
}

type UsageInMinutesRequest struct {
	ServerId int `json:"server_id" binding:"required"`

	From   *time.Time           `json:"date_from"`
	To     *time.Time           `json:"date_to"`
	Period analytics.TimePeriod `json:"period" binding:"required"`
}

type UsageInMinutesResponse struct {
	Items []*models.PeriodData `json:"items"`
}

type PlayersOnlineRequest struct {
	ServerId int `json:"server_id"`

	From   *time.Time           `json:"date_from"`
	To     *time.Time           `json:"date_to"`
	Period analytics.TimePeriod `json:"period" binding:"required"`
}

type PlayersOnlineResponse struct {
	Items []*models.PeriodData `json:"items"`
}

type PopularServersResponseItem struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Difficulty int    `json:"diff"`

	TotalSessions int `json:"total_sessions"`
	TotalUsers    int `json:"total_users"`
}

type PopularServersResponse struct {
	Items []*PopularServersResponseItem `json:"items"`
}

type TotalOnlineResponse struct {
	Count int `json:"count"`
}

type SessionCountHistRequest struct {
	ServerId int `json:"server_id" binding:"required"`

	From *time.Time `json:"date_from"`
	To   *time.Time `json:"date_to"`

	MapIds []int `json:"map_ids"`

	Statuses []models.GameStatus `json:"statuses"`

	SpawnCycle  *string `json:"spawn_cycle"`
	ZedsType    *string `json:"zeds_type"`
	MinWave     *int    `json:"min_wave"`
	MaxMonsters *int    `json:"max_monsters"`

	AuthUser *models.TokenPayload `json:"-"`
}
