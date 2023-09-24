package models

import "time"

type AuthType int

const (
	Steam AuthType = iota
	EGS
)

type User struct {
	Id     int
	AuthId string
	Type   AuthType

	Name string
}

type UserNameHistory struct {
	UserId int
	Name   string

	UpdatedAt time.Time
}
