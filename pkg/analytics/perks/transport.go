package perks

import "time"

type PerkStats struct {
	Perk  int `json:"perk"`
	Count int `json:"count"`
}

type PerksPlayTimeRequest struct {
	ServerId int `json:"server_id"`
	UserId   int `json:"user_id"`

	From time.Time `json:"date_from" binding:"required"`
	To   time.Time `json:"date_to" binding:"required"`
}

type PerksPlayTimeResponse struct {
	Items []PerkStats `json:"items"`
}

type PerksKillsRequest struct {
	ServerId int `json:"server_id"`
	UserId   int `json:"user_id"`

	From time.Time `json:"date_from" binding:"required"`
	To   time.Time `json:"date_to" binding:"required"`
}

type PerksKillsResponse struct {
	Items []PerkStats `json:"items"`
}
