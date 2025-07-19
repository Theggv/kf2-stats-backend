package models

const TokenVersion = 1

type TokenPayload struct {
	UserId int    `json:"user_id"`
	Name   string `json:"name"`

	SteamId    string `json:"steam_id"`
	Avatar     string `json:"avatar"`
	ProfileUrl string `json:"profile_url"`
}
