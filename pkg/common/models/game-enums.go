package models

type AuthType int

const (
	Steam AuthType = iota + 1
	EGS
)

type Perk = int

const (
	Berserker Perk = iota + 1
	Commando
	Medic
	Sharpshooter
	Gunslinger
	Support
	Swat
	Demolitionist
	Firebug
	Survivalist
)

type GameMode = int

const (
	Any GameMode = iota
	Survival
	Endless
	ControlledDifficulty
	Weekly
	Objective
	Versus
)

type GameLength = int

const (
	Short  GameLength = 4
	Medium GameLength = 7
	Long   GameLength = 10
	Custom GameLength = -1
)

type GameDifficulty = int

const (
	Normal GameDifficulty = iota + 1
	Hard
	Suicidal
	HellOnEarth
)

type GameStatus = int

const (
	Lobby GameStatus = iota
	InProgress
	Win
	Lose
	Solomode
	Aborted GameStatus = -1
)
