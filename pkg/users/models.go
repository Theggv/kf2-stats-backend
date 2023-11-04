package users

import "github.com/theggv/kf2-stats-backend/pkg/common/models"

type User struct {
	Id     int             `json:"id"`
	AuthId string          `json:"auth_id"`
	Type   models.AuthType `json:"auth_type"`

	Name string `json:"name"`
}
