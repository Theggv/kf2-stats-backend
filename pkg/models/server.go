package models

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

type ServerType = int

const (
	Vanilla ServerType = iota
	HellOnEarthPlus
	ControlledDifficulty
)

type Server struct {
	Id      int
	Name    string
	Address string
	Type    ServerType
}

type ServerSession struct {
	Id       int
	ServerId int
	Mode     Mode
	Length   Length
}
