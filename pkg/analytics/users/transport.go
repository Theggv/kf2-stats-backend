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
	AuthUser *models.TokenPayload `json:"-"`
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
	UserId int `json:"user_id" binding:"required"`

	Pager models.PaginationRequest `json:"pager"`

	AuthUser *models.TokenPayload `json:"-"`
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
	Items    []*GetTeammatesResponseItem `json:"items"`
	Metadata *models.PaginationResponse  `json:"metadata"`
}

type GetPlayedMapsRequest struct {
	UserId int `json:"user_id" binding:"required"`

	Perks     []int `json:"perks"`
	ServerIds []int `json:"server_ids"`

	From *time.Time `json:"date_from"`
	To   *time.Time `json:"date_to"`
}

type GetPlayedMapsResponseItem struct {
	Name string `json:"name"`

	TotalGames int `json:"total_games"`
	TotalWins  int `json:"total_wins"`

	LastPlayed *time.Time `json:"last_played"`
}

type GetPlayedMapsResponse struct {
	Items []*GetPlayedMapsResponseItem `json:"items"`
}

type GetLastSeenUsersRequest struct {
	UserId     int    `json:"user_id" binding:"required"`
	SearchText string `json:"search_text"`

	Perks     []int `json:"perks"`
	ServerIds []int `json:"server_ids"`

	From *time.Time `json:"date_from"`
	To   *time.Time `json:"date_to"`

	Pager models.PaginationRequest `json:"pager"`
}

type SessionData struct {
	Id int `json:"id"`

	Mode       models.GameMode       `json:"mode"`
	Length     int                   `json:"length"`
	Difficulty models.GameDifficulty `json:"diff"`

	Status models.GameStatus `json:"status"`
}

type ServerData struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type MapData struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type GetLastSeenUsersResponseItem struct {
	Id   int    `json:"id"`
	Name string `json:"name"`

	ProfileUrl *string `json:"profile_url"`
	Avatar     *string `json:"avatar"`

	Session SessionData `json:"session"`
	Server  ServerData  `json:"server"`
	Map     MapData     `json:"map"`

	Perks []int `json:"perks"`

	LastSeen *time.Time `json:"last_seen"`

	AuthId string          `json:"-"`
	Type   models.AuthType `json:"-"`
}

type GetLastSeenUsersResponse struct {
	Items    []*GetLastSeenUsersResponseItem `json:"items"`
	Metadata *models.PaginationResponse      `json:"metadata"`
}

type GetLastSessionsWithUserRequest struct {
	UserId      int `json:"user_id" binding:"required"`
	OtherUserId int `json:"other_user_id" binding:"required"`

	Perks     []int `json:"perks"`
	ServerIds []int `json:"server_ids"`

	From *time.Time `json:"date_from"`
	To   *time.Time `json:"date_to"`

	Pager models.PaginationRequest `json:"pager"`
}

type GetLastSessionsWithUserResponseItem struct {
	Session SessionData `json:"session"`
	Server  ServerData  `json:"server"`
	Map     MapData     `json:"map"`

	Perks []int `json:"perks"`

	LastSeen *time.Time `json:"last_seen"`
}

type GetLastSessionsWithUserResponse struct {
	Items    []*GetLastSessionsWithUserResponseItem `json:"items"`
	Metadata *models.PaginationResponse             `json:"metadata"`
}
