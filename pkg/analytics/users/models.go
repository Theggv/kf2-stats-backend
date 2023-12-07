package users

type UsersTopType = int

const (
	MostKills UsersTopType = iota
	MostDeaths
	MostPlaytime
	MostDamageDealt
	MostHsAccuracy
	MostHealsGiven
	AverageZedtime
)
