package users

import (
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

type CreateUserRequest struct {
	AuthId   string   `json:"auth_id"`
	AuthType AuthType `json:"auth_type"`

	Name string `json:"name"`
}

type CreateUserResponse struct {
	Id int `json:"id"`
}

type FilterUsersRequest struct {
	Pager models.PaginationRequest `json:"pager"`
}

type FilterUsersResponseUserSession struct {
	Id int `json:"id"`

	Mode       models.GameMode       `json:"mode"`
	Length     int                   `json:"length"`
	Difficulty models.GameDifficulty `json:"diff"`

	Status models.GameStatus `json:"status"`

	CDData *models.CDGameData `json:"cd_data"`

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

	AuthId string   `json:"-"`
	Type   AuthType `json:"-"`

	LastSessionId    *int `json:"-"`
	CurrentSessionId *int `json:"-"`
}

type FilterUsersResponse struct {
	Items    []*FilterUsersResponseUser `json:"items"`
	Metadata models.PaginationResponse  `json:"metadata"`
}
