package demorecord

import (
	"math"
)

type DifficultyScore struct {
	Score float64 `json:"score"`

	WaveSizeBonus float64 `json:"wave_size_bonus"`
	SpeedBonus    float64 `json:"speed_bonus"`
	ZedsBonus     float64 `json:"zeds_bonus"`
	PlayersBonus  float64 `json:"players_bonus"`
	ZedtimeBonus  float64 `json:"zt_bonus"`
}

type DifficultyAnalyticsDetails struct {
	Step   int `json:"step"`
	Period int `json:"period"`

	Buckets []*DifficultyScore `json:"buckets"`
}

type ZedsSnapshot struct {
	Tick  int `json:"tick"`
	Score int `json:"score"`
}

type DifficultyAnalytics struct {
	OverAll *DifficultyScore `json:"overall"`

	Details      *DifficultyAnalyticsDetails `json:"details,omitempty"`
	Distribution []*ZedsSnapshot             `json:"zeds_distribution"`
}

func calcWaveSizeBonus(waveSize int) float64 {
	if waveSize > 20 {
		return lerp(0.1, 1, float64(min(waveSize, 411))/float64(411))
	}

	return 0.1
}

func calcPlayerBonus(numPlayers int) float64 {
	playerDifficultyMultipliers := []float64{0, 3.5, 2.75, 2.25, 1.6, 1.25, 1}

	if numPlayers < len(playerDifficultyMultipliers) {
		return playerDifficultyMultipliers[numPlayers]
	}

	return lerp(1, 0.5, min(6.0, float64(numPlayers)-6.0)/6.0)
}

func calcDifficultyZedsBonus(counter *ZedCounter) float64 {
	if counter.Total <= 0 {
		return 0
	}

	trashMp := 1.0
	mediumMp := 2.0
	largeMp := 5.0

	bonus := (float64(counter.Trash)*trashMp +
		float64(counter.Medium)*mediumMp +
		float64(counter.Large)*largeMp) /
		float64(counter.Total)

	// [0; 1]
	return max(0, (bonus-1.0)/4.0)
}

func calcZtBonus(zedTimes []*DemoRecordAnalysisZedtime, tick, period int) float64 {
	startTick := tick - period
	endTick := tick - 1
	duration := float64(endTick-startTick) / 100

	maxMultiplier := 4.0

	for i := range zedTimes {
		item := zedTimes[i]

		ztStartTick := item.MetaData.StartTick
		ztEndTick := item.MetaData.EndTick + 2000

		if startTick >= ztStartTick && startTick <= ztEndTick ||
			endTick >= ztStartTick && endTick <= ztEndTick {

			// period intersects with zedtime
			if startTick >= ztStartTick && endTick <= ztEndTick {
				// period inside zedtime
				return maxMultiplier
			} else if startTick >= ztStartTick {
				// zedtime started before period
				return lerp(1, maxMultiplier, float64(ztEndTick-startTick)/100/duration)
			} else {
				// zedtime started after period
				return lerp(1, maxMultiplier, float64(endTick-ztStartTick)/100/duration)
			}
		}

		if ztStartTick >= startTick && ztEndTick <= endTick {
			// zedtime inside period
			return lerp(1, maxMultiplier, item.MetaData.Duration/duration)
		}
	}

	return 1.0
}

func (wave *DemoRecordAnalysisWave) calcZedsDistribution() []*ZedsSnapshot {
	getZedWeight := func(kill *DemoRecordParsedEventKill) int {
		if kill.IsLarge() {
			return 5
		} else if kill.IsMedium() {
			return 2
		} else if kill.IsTrash() {
			return 1
		}

		return 1
	}

	res := []*ZedsSnapshot{}

	waveSize := len(wave.PlayerEvents.Kills)
	maxMonsters := min(40, waveSize)

	score := 0

	for i := range maxMonsters {
		kill := wave.PlayerEvents.Kills[i]

		score += getZedWeight(kill)
	}

	res = append(res, &ZedsSnapshot{
		Tick:  wave.MetaData.StartTick,
		Score: score,
	})

	for i := range waveSize {
		kill := wave.PlayerEvents.Kills[i]

		res = append(res, &ZedsSnapshot{
			Tick:  kill.Tick,
			Score: score,
		})

		if i+maxMonsters < waveSize {
			score -= getZedWeight(kill)
			score += getZedWeight(wave.PlayerEvents.Kills[i+maxMonsters])
		}
	}

	res = append(res, &ZedsSnapshot{
		Tick:  wave.MetaData.EndTick,
		Score: res[len(res)-1].Score,
	})

	return res

}

func (wave *DemoRecordAnalysisWave) calcDifficulty(step, period int) {
	res := DifficultyAnalytics{
		OverAll: &DifficultyScore{},
		Details: &DifficultyAnalyticsDetails{
			Step:    step,
			Period:  period,
			Buckets: []*DifficultyScore{},
		},
		Distribution: wave.calcZedsDistribution(),
	}

	summary := wave.Analytics.Summary

	waveStartTick := wave.MetaData.StartTick
	waveEndTick := wave.MetaData.EndTick

	waveSizeBonus := calcWaveSizeBonus(summary.WaveSize)
	playersBonus := calcPlayerBonus(len(wave.PlayerEvents.Perks))

	var meanZedDistr float64 = 1
	if len(res.Distribution) > 0 {
		sum := 0
		for i := range res.Distribution {
			sum += res.Distribution[i].Score
		}
		meanZedDistr = float64(sum) / float64(len(res.Distribution))
	}

	// offset represents end tick of period
	// start tick = offset - period
	for offset := waveStartTick; offset <= waveEndTick+step; offset += step {
		counter := ZedCounter{}

		kills := filterByRange(
			wave.PlayerEvents.Kills,
			func(item *DemoRecordParsedEventKill) int {
				return item.Tick
			}, offset-period, offset-1,
		)

		zedsDistrib := findLastLower(
			res.Distribution,
			func(item *ZedsSnapshot) int {
				return item.Tick
			}, max(offset-period, waveStartTick),
		)

		for i := range kills {
			kill := kills[i]

			if kill.IsLarge() {
				counter.Large += 1
			} else if kill.IsMedium() {
				counter.Medium += 1
			} else if kill.IsTrash() {
				counter.Trash += 1
			}

			counter.Total += 1
		}

		var avgZedsPerSecond float64
		{
			kills := filterByRange(
				wave.PlayerEvents.Kills,
				func(item *DemoRecordParsedEventKill) int {
					return item.Tick
				}, offset-3000, offset-1,
			)

			duration := float64(period) / 100
			avgZedsPerSecond = float64(len(kills)) / duration
		}

		zedsDistributionBonus := float64((*zedsDistrib).Score) / max(1, meanZedDistr)
		ztBonus := calcZtBonus(wave.Zedtimes, offset, period)
		zedsBonus := calcDifficultyZedsBonus(&counter) * waveSizeBonus * zedsDistributionBonus
		speedBonus := avgZedsPerSecond * playersBonus * lerp(0.1, 1, zedsBonus)

		bucket := DifficultyScore{
			Score:         1*(1+zedsBonus) + 1.5*speedBonus*ztBonus,
			WaveSizeBonus: waveSizeBonus,
			SpeedBonus:    speedBonus,
			ZedsBonus:     zedsBonus,
			PlayersBonus:  playersBonus,
			ZedtimeBonus:  ztBonus,
		}

		res.Details.Buckets = append(res.Details.Buckets, &bucket)
	}

	var weights []float64
	weightsCount := 15
	for i := range weightsCount {
		weights = append(weights, math.Exp(-float64(i)/float64(weightsCount)))
	}

	for i := len(res.Details.Buckets) - 1; i > 0; i-- {
		for j := i - 1; j >= max(0, i-weightsCount+1); j-- {
			res.Details.Buckets[i].Score += weights[i-j] * res.Details.Buckets[j].Score
		}
	}

	wave.Analytics.Difficulty = &res
}
