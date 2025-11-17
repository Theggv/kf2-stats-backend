package models

type UserProfile struct {
	Id int `json:"id"`

	Name       string  `json:"name"`
	ProfileUrl *string `json:"profile_url"`
	Avatar     *string `json:"avatar"`

	Type   AuthType `json:"-"`
	AuthId string   `json:"-"`
}

func (s *UserProfile) AsFull() *UserProfileFull {
	return &UserProfileFull{
		Id:         s.Id,
		Name:       s.Name,
		ProfileUrl: s.ProfileUrl,
		Avatar:     s.Avatar,
		Type:       s.Type,
		AuthId:     s.AuthId,
	}
}

type UserProfileFull struct {
	Id int `json:"id"`

	Name       string  `json:"name"`
	ProfileUrl *string `json:"profile_url"`
	Avatar     *string `json:"avatar"`

	Type   AuthType `json:"auth_type"`
	AuthId string   `json:"auth_id"`
}

func (s *UserProfileFull) AsPartial() *UserProfile {
	return &UserProfile{
		Id:         s.Id,
		Name:       s.Name,
		ProfileUrl: s.ProfileUrl,
		Avatar:     s.Avatar,
		Type:       s.Type,
		AuthId:     s.AuthId,
	}
}
