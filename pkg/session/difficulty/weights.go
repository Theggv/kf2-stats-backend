package difficulty

import (
	"math"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

func calcTotalPlayersWeight(totalPlayers int) float64 {
	totalPlayerWeights := []float64{1.7, 1.35, 1.25, 1.15, 1.10, 1}

	if totalPlayers < 0 {
		return 0
	}

	if totalPlayers <= 6 {
		return totalPlayerWeights[totalPlayers-1]
	}

	return npInterp(float64(totalPlayers-6)/6.0, pair{0, 1}, pair{1, 0.5})
}

func calcWaveWeight(wave int) float64 {
	if wave == 1 {
		return 1.2
	}

	if wave == 2 {
		return 1.05
	}

	return 1
}

func calcZedsWeight(data map[string]int) float64 {
	weights := map[string]float64{
		"cyst":          1,
		"alpha_clot":    1,
		"slasher":       1,
		"stalker":       1,
		"crawler":       1,
		"gorefast":      1,
		"elite_crawler": 1,
		"rioter":        1,
		"gorefiend":     1,
		"siren":         1,

		"bloat": 2,
		"husk":  2,
		"edar":  2,

		"scrake": 4,
		"fp":     5,
		"qp":     4,
	}

	result := 0.0

	for key, count := range data {
		if mp, ok := weights[key]; ok {
			result += mp * float64(count)
		}
	}

	return result
}

func calcTotalZeds(data map[string]int) int {
	result := 0

	for _, val := range data {
		result += val
	}

	return result
}

func calcWaveZedsDifficulty(
	zedsType string,
	wave int,
	gameLength models.GameLength,
	gameDifficulty models.GameDifficulty,
	data map[string]int,
) float64 {
	totalZeds := calcTotalZeds(data)
	if totalZeds <= 0 {
		return 0
	}

	zedWeigts := map[string]map[string]float64{
		"vanilla": {
			"cyst": 90, "alpha_clot": 115, "slasher": 100, "gorefast": 140,
			"stalker": 150, "crawler": 150, "elite_crawler": 170,

			"rioter": 190, "gorefiend": 250, "edar": 250,
			"siren": 220, "bloat": 200, "husk": 250,
			"scrake": 800, "fp": 1200, "qp": 900,
		},
		"harder": {
			"cyst": 100, "alpha_clot": 120, "slasher": 110, "gorefast": 160,
			"stalker": 160, "crawler": 160, "elite_crawler": 170,

			"rioter": 190, "gorefiend": 250, "edar": 250,
			"siren": 300, "bloat": 220, "husk": 300,
			"scrake": 1200, "fp": 1500, "qp": 900,
		},
		"nightcore": {
			"cyst": 105, "alpha_clot": 125, "slasher": 115, "gorefast": 165,
			"stalker": 165, "crawler": 165, "elite_crawler": 175,

			"rioter": 190, "gorefiend": 250, "edar": 250,
			"siren": 350, "bloat": 225, "husk": 320,
			"scrake": 1300, "fp": 1600, "qp": 900,
		},
	}

	gameLengthWaveWeights := map[int]map[models.GameLength]float64{
		models.Short:  {1: 1.25, 2: 1.05},
		models.Medium: {1: 1.25, 2: 1.1},
		models.Long:   {1: 1.25, 2: 1.1},
	}

	gameDifficultyWeights := map[models.GameDifficulty]float64{
		models.Normal:      0.5,
		models.Hard:        0.7,
		models.Suicidal:    0.85,
		models.HellOnEarth: 1,
	}

	zedsTypeWeights, ok := zedWeigts[zedsType]
	if !ok {
		zedsTypeWeights = zedWeigts["vanilla"]
	}

	result := 0.0

	for key, count := range data {
		if mp, ok := zedsTypeWeights[key]; ok {
			result += mp * float64(count)
		}
	}

	result = result / float64(totalZeds) / 100

	if waveWeights, ok := gameLengthWaveWeights[gameLength]; ok {
		if value, ok := waveWeights[wave]; ok {
			result = result * value
		}
	}

	if value, ok := gameDifficultyWeights[gameDifficulty]; ok {
		result = result * value
	}

	return result
}

func predictDuration(
	zeds map[string]int,
	totalPlayers int,
	wave int,
) float64 {
	zedsWeight := calcZedsWeight(zeds)
	totalPlayersMp := calcTotalPlayersWeight(totalPlayers)
	waveMp := calcWaveWeight(wave)

	pred := 67.1369 + 0.2548*(zedsWeight*totalPlayersMp*waveMp)

	return pred
}

func calcKitingPenalty(duration float64, expectedDuration float64) float64 {
	// predicted: 240s
	// eps = 0.125 * 240s = 30s
	// 3 * eps = 90s
	// expected duration should be in [150s; 330s]
	//
	// (360s) duration/predicted = 1.5 -> 1.5 - 1 - 0.125 * 3 = 0.125 -> 0.3 -> penalty = 0.81
	// (480s) duration/predicted = 2 -> 2 - 1 - 0.125 * 3 = 0.625 -> 1.5 -> penalty = 0.35
	// (630+s) -> max penalty (0.125)

	eps := 0.125
	exp := npInterp((duration/expectedDuration)-1-eps*3, pair{0, 1.25}, pair{0, 3})

	return 1 / math.Pow(2, exp)
}

func calcWaveSizePenalty(totalZeds int) float64 {
	return npInterp(float64(totalZeds), pair{0, 100}, pair{0.25, 1})
}

func calcTotalPlayersPenalty(totalPlayers int) float64 {
	return npInterp(float64(totalPlayers-6)/6.0, pair{0, 1}, pair{1, 0.5})
}
