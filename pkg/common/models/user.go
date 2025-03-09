package models

type UserProfile struct {
	Id     int    `json:"id"`
	AuthId string `json:"auth_id"`

	Name       string  `json:"name"`
	ProfileUrl *string `json:"profile_url"`
	Avatar     *string `json:"avatar"`
}
