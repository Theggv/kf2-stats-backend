package leaderboards

type LeaderBoardType = int

const (
	TotalGames LeaderBoardType = iota + 1
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
