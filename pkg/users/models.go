package users

import "time"

type AuthType int

const (
	Steam AuthType = iota + 1
	EGS
)

type User struct {
	Id     int      `json:"id"`
	AuthId string   `json:"auth_id"`
	Type   AuthType `json:"auth_type"`

	Name string `json:"name"`
}

type UserNameHistory struct {
	UserId int    `json:"user_id"`
	Name   string `json:"name"`

	UpdatedAt time.Time `json:"updated_at"`
}
