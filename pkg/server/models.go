package server

type ServerType = int

const (
	Vanilla ServerType = iota
	HellOnEarthPlus
	ControlledDifficulty
)

type Server struct {
	Id      int        `json:"id"`
	Name    string     `json:"name"`
	Address string     `json:"address"`
	Type    ServerType `json:"type"`
}
