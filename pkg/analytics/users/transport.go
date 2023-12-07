package users

import (
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

type UserAnalyticsRequest struct {
	UserId int `json:"user_id"`

	From *time.Time `json:"date_from"`
	To   *time.Time `json:"date_to"`
}

type UserAnalyticsResponse struct {
	Games int `json:"games"`
	Wins  int `json:"wins"`

	Minutes int `json:"minutes"`

	Kills  int `json:"kills"`
	Deaths int `json:"deaths"`
}

type UserPerksAnalyticsRequest struct {
	UserId int `json:"user_id"`

	From *time.Time `json:"date_from"`
	To   *time.Time `json:"date_to"`
}

type UserPerksAnalyticsResponseItem struct {
	Perk int `json:"perk"`

	Games int `json:"games"`
	Wins  int `json:"wins"`

	Kills      int `json:"kills"`
	LargeKills int `json:"large_kills"`

	Waves  int `json:"waves"`
	Deaths int `json:"deaths"`

	Accuracy   float64 `json:"accuracy"`
	HSAccuracy float64 `json:"hs_accuracy"`

	DamageDealt int `json:"damage_dealt"`
	DamageTaken int `json:"damage_taken"`
	HealsGiven  int `json:"heals_given"`

	Minutes int `json:"minutes"`
}

type UserPerksAnalyticsResponse struct {
	AverageZedtime float64 `json:"avg_zt"`

	Items []UserPerksAnalyticsResponseItem `json:"items"`
}

type UserPerkHistRequest struct {
	UserId int  `json:"user_id" binding:"required"`
	Perk   *int `json:"perk"`

	From *time.Time `json:"date_from"`
	To   *time.Time `json:"date_to"`
}

type AccuracyHistItem struct {
	Period time.Time `json:"period"`

	Accuracy   float64 `json:"accuracy"`
	HSAccuracy float64 `json:"hs_accuracy"`
}

type AccuracyHist struct {
	Items []AccuracyHistItem `json:"items"`
}

type PlayTimeHistItem struct {
	Period time.Time `json:"period"`

	Count   int `json:"count"`
	Minutes int `json:"minutes"`
}

type PlayTimeHist struct {
	Items []PlayTimeHistItem `json:"items"`
}

type GetTeammatesRequest struct {
	UserId int  `json:"user_id" binding:"required"`
	Limit  *int `json:"limit"`
}

type GetTeammatesResponseItem struct {
	Id   int    `json:"id"`
	Name string `json:"name"`

	ProfileUrl *string `json:"profile_url"`
	Avatar     *string `json:"avatar"`

	Games int `json:"games"`
	Wins  int `json:"wins"`

	AuthId string          `json:"-"`
	Type   models.AuthType `json:"-"`
}

type GetTeammatesResponse struct {
	Items []*GetTeammatesResponseItem `json:"items"`
}

type GetUsersTopRequest struct {
	Type UsersTopType `json:"type" binding:"required"`
	Perk int          `json:"perk"`

	From *time.Time `json:"date_from" binding:"required"`
	To   *time.Time `json:"date_to" binding:"required"`
}

type GetUsersTopResponseItem struct {
	Id   int    `json:"id"`
	Name string `json:"name"`

	ProfileUrl *string `json:"profile_url"`
	Avatar     *string `json:"avatar"`

	Games  int     `json:"games"`
	Metric float64 `json:"metric"`

	AuthId string          `json:"-"`
	Type   models.AuthType `json:"-"`
}

type GetUsersTopResponse struct {
	Items []*GetUsersTopResponseItem `json:"items"`
}
