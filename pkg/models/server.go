package models

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
