package users

import "time"

type UserAnalyticsRequest struct {
	UserId int `json:"user_id"`

	From time.Time `json:"date_from"`
	To   time.Time `json:"date_to"`
}

type UserAnalyticsResponse struct {
	Games int `json:"games"`
	Wins  int `json:"wins"`

	Minutes int `json:"minutes"`

	Kills  int `json:"kills"`
	Deaths int `json:"deaths"`
}
