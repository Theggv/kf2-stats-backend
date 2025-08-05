package users

import (
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

type CreateUserRequest struct {
	AuthId   string          `json:"auth_id"`
	AuthType models.AuthType `json:"auth_type"`

	Name string `json:"name"`
}

type CreateUserResponse struct {
	Id int `json:"id"`
}

type FilterUsersRequest struct {
	SearchText string `json:"search_text"`

	Pager models.PaginationRequest `json:"pager"`
}

type FilterUsersResponseUserSession struct {
	Id int `json:"id"`

	Mode       models.GameMode       `json:"mode"`
	Length     int                   `json:"length"`
	Difficulty models.GameDifficulty `json:"diff"`

	Status models.GameStatus `json:"status"`

	Wave   int                   `json:"wave"`
	CDData *models.ExtraGameData `json:"cd_data"`

	ServerName string `json:"server_name"`
	MapName    string `json:"map_name"`
}

type FilterUsersResponseUser struct {
	Id   int    `json:"id"`
	Name string `json:"name"`

	ProfileUrl *string `json:"profile_url"`
	Avatar     *string `json:"avatar"`

	LastSession    *FilterUsersResponseUserSession `json:"last_session"`
	CurrentSession *FilterUsersResponseUserSession `json:"current_session"`

	UpdatedAt *time.Time `json:"updated_at"`

	AuthId string          `json:"-"`
	Type   models.AuthType `json:"-"`

	LastSessionId    *int `json:"-"`
	CurrentSessionId *int `json:"-"`
}

type FilterUsersResponse struct {
	Items    []*FilterUsersResponseUser `json:"items"`
	Metadata models.PaginationResponse  `json:"metadata"`
}

type RecentSessionsRequest struct {
	UserId int `json:"user_id" binding:"required"`

	Date *time.Time `json:"date"`

	Perks     []int `json:"perks"`
	ServerIds []int `json:"server_ids"`
	MapIds    []int `json:"map_ids"`

	Mode       *models.GameMode       `json:"mode"`
	Length     *models.GameLength     `json:"length"`
	Difficulty *models.GameDifficulty `json:"diff"`
	Status     *models.GameStatus     `json:"status"`

	SpawnCycle     *string `json:"spawn_cycle"`
	MinWave        *int    `json:"min_wave"`
	MinMaxMonsters *int    `json:"min_mm"`

	Pager models.PaginationRequest `json:"pager"`
}

type RecentSessionsResponseSessionServer struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type RecentSessionsResponseSession struct {
	Id int `json:"id"`

	Mode       models.GameMode       `json:"mode"`
	Length     int                   `json:"length"`
	Difficulty models.GameDifficulty `json:"diff"`

	Status models.GameStatus `json:"status"`

	Wave   int                   `json:"wave"`
	CDData *models.ExtraGameData `json:"cd_data"`

	MapName string                              `json:"map_name"`
	Server  RecentSessionsResponseSessionServer `json:"server"`

	Perks []int `json:"perks"`

	UpdatedAt *time.Time `json:"updated_at"`
}

type RecentSessionsResponse struct {
	Items    []*RecentSessionsResponseSession `json:"items"`
	Metadata models.PaginationResponse        `json:"metadata"`
}
