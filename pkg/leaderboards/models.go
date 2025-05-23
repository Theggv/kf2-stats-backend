package leaderboards

type LeaderBoardOrderBy = int

const (
	TotalGames LeaderBoardOrderBy = iota + 1
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
	AverageBuffsUptime

	TotalPlaytime
)
