package users

type UsersTopType = int

const (
	TotalGames UsersTopType = iota + 1
	TotalDeaths

	Accuracy
	HsAccuracy

	TotalDamage
	MostDamage

	TotalKills
	TotalLargeKills
	TotalHuskRages

	TotalHeals

	AverageZedtime
	TotalPlaytime
)
