package users

type CreateUserRequest struct {
	AuthId string   `json:"auth_id"`
	Type   AuthType `json:"auth_type"`

	Name string `json:"name"`
}

type CreateUserResponse struct {
	Id int `json:"id"`
}
