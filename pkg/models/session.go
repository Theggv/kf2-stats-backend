package models

import "time"

type Mode = int

const (
	Survival Mode = iota
	Endless
	Any Mode = -1
)

type Length = int

const (
	Short Length = iota
	Medium
	Long
	NotSupported Length = -1
)

type Difficulty = int

const (
	Normal Difficulty = iota
	Hard
	Suicidal
	HellOnEarth
)

type Status = int

const (
	InGame Status = iota
	Win
	Lose
	Unknown Status = -1
)

type Session struct {
	Id       int
	ServerId int
	MapId    int

	Mode       Mode
	Length     Length
	Difficulty Difficulty

	Status Status

	CreatedAt time.Time
	UpdatedAt time.Time
}
