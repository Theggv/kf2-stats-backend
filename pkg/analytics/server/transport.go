package server

import "time"

type TimePeriod = int

const (
	Hour TimePeriod = iota + 1
	Day
	Week
	Month
	Year
)

type SessionCountRequest struct {
	ServerId int `json:"server_id" binding:"required"`

	From   time.Time  `json:"date_from" binding:"required"`
	To     time.Time  `json:"date_to" binding:"required"`
	Period TimePeriod `json:"period" binding:"required"`
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

	From   time.Time  `json:"date_from" binding:"required"`
	To     time.Time  `json:"date_to" binding:"required"`
	Period TimePeriod `json:"period" binding:"required"`
}

type UsageInMinutesResponse struct {
	Items []PeriodData `json:"items"`
}

type PlayersOnlineRequest struct {
	ServerId int `json:"server_id" binding:"required"`

	From   time.Time  `json:"date_from" binding:"required"`
	To     time.Time  `json:"date_to" binding:"required"`
	Period TimePeriod `json:"period" binding:"required"`
}

type PlayersOnlineResponse struct {
	Items []PeriodData `json:"items"`
}

type PerkStats struct {
	Perk  int `json:"perk"`
	Count int `json:"count"`
}

type PerksPlayTimeRequest struct {
	ServerId int `json:"server_id"`

	From time.Time `json:"date_from" binding:"required"`
	To   time.Time `json:"date_to" binding:"required"`
}

type PerksPlayTimeResponse struct {
	Items []PerkStats `json:"items"`
}
