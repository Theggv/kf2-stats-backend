package session

import "time"

type Mode = int

const (
	Any Mode = iota
	Survival
	Endless
	ControlledDifficulty
)

type Length = int

const (
	Short Length = iota + 1
	Medium
	Long
	NotSupported Length = -1
)

type Difficulty = int

const (
	Normal Difficulty = iota + 1
	Hard
	Suicidal
	HellOnEarth
)

type Status = int

const (
	Lobby Status = iota
	InProgress
	Win
	Lose
	Solomode
	Aborted Status = -1
)

type SessionMap struct {
	Name    *string `json:"name"`
	Preview *string `json:"preview"`
}

type SessionServer struct {
	Name    *string `json:"name"`
	Address *string `json:"address"`
}

type Session struct {
	Id       int `json:"id"`
	ServerId int `json:"server_id"`
	MapId    int `json:"map_id"`

	Mode       Mode       `json:"mode"`
	Length     Length     `json:"length"`
	Difficulty Difficulty `json:"diff"`

	Status Status `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Map    *SessionMap    `json:"map"`
	Server *SessionServer `json:"server"`
}
