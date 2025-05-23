package leaderboards

import (
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

type LeaderBoardsRequest struct {
	ServerIds []int `json:"server_id"`

	Perk int `json:"perk"`

	From time.Time `json:"date_from" binding:"required"`
	To   time.Time `json:"date_to" binding:"required"`

	OrderBy LeaderBoardOrderBy `json:"type" binding:"required"`
	Page    int                `json:"page"`
}

type MostDamageMatch struct {
	SessionId int     `json:"session_id"`
	Value     float64 `json:"value"`
}

type LeaderBoardsResponseItem struct {
	Id   int    `json:"id"`
	Name string `json:"name"`

	ProfileUrl *string `json:"profile_url"`
	Avatar     *string `json:"avatar"`

	TotalGames  int `json:"total_games"`
	TotalDeaths int `json:"total_deaths"`

	Accuracy   float64 `json:"accuracy"`
	HSAccuracy float64 `json:"hs_accuracy"`

	TotalDamage int              `json:"total_damage"`
	MostDamage  *MostDamageMatch `json:"most_damage"`

	TotalKills      int `json:"total_kills"`
	TotalLargeKills int `json:"total_large_kills"`
	TotalHuskRages  int `json:"total_husk_rages"`

	TotalHeals     int     `json:"total_heals"`
	AverageZedtime float64 `json:"avg_zt"`

	TotalPlaytime int `json:"total_playtime"`

	AuthId string          `json:"-"`
	Type   models.AuthType `json:"-"`
}

type LeaderBoardsResponse struct {
	Items    []*LeaderBoardsResponseItem `json:"items"`
	Metadata *models.PaginationResponse  `json:"metadata"`
}
