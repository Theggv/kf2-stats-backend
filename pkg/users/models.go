package users

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
