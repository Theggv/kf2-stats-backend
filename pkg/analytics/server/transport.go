package server

import (
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/analytics"
)

type SessionCountRequest struct {
	ServerId int `json:"server_id" binding:"required"`

	From   time.Time            `json:"date_from" binding:"required"`
	To     time.Time            `json:"date_to" binding:"required"`
	Period analytics.TimePeriod `json:"period" binding:"required"`
}

type PeriodData struct {
	Count  int    `json:"count"`
	Period string `json:"period"`
}

type SessionCountResponse struct {
	Items []PeriodData `json:"items"`
}

type UsageInMinutesRequest struct {
	ServerId int `json:"server_id" binding:"required"`

	From   time.Time            `json:"date_from" binding:"required"`
	To     time.Time            `json:"date_to" binding:"required"`
	Period analytics.TimePeriod `json:"period" binding:"required"`
}

type UsageInMinutesResponse struct {
	Items []PeriodData `json:"items"`
}

type PlayersOnlineRequest struct {
	ServerId int `json:"server_id" binding:"required"`

	From   time.Time            `json:"date_from" binding:"required"`
	To     time.Time            `json:"date_to" binding:"required"`
	Period analytics.TimePeriod `json:"period" binding:"required"`
}

type PlayersOnlineResponse struct {
	Items []PeriodData `json:"items"`
}
