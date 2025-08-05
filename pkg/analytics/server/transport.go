package server

import (
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/analytics"
)

type SessionCountRequest struct {
	ServerId int `json:"server_id"`

	From   *time.Time           `json:"date_from"`
	To     *time.Time           `json:"date_to"`
	Period analytics.TimePeriod `json:"period" binding:"required"`
}

type PeriodData struct {
	Period string `json:"period"`

	Value         int     `json:"value"`
	PreviousValue int     `json:"prev"`
	Difference    int     `json:"diff"`
	MaxValue      int     `json:"max_value"`
	Trend         float64 `json:"trend"`
}

type SessionCountResponse struct {
	Items []*PeriodData `json:"items"`
}

type UsageInMinutesRequest struct {
	ServerId int `json:"server_id" binding:"required"`

	From   *time.Time           `json:"date_from"`
	To     *time.Time           `json:"date_to"`
	Period analytics.TimePeriod `json:"period" binding:"required"`
}

type UsageInMinutesResponse struct {
	Items []*PeriodData `json:"items"`
}

type PlayersOnlineRequest struct {
	ServerId int `json:"server_id"`

	From   *time.Time           `json:"date_from"`
	To     *time.Time           `json:"date_to"`
	Period analytics.TimePeriod `json:"period" binding:"required"`
}

type PlayersOnlineResponse struct {
	Items []*PeriodData `json:"items"`
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
