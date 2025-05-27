package demorecord

import "github.com/theggv/kf2-stats-backend/pkg/common/models"

func (wave *DemoRecordAnalysisWave) calcZedtimeAnalytics() *ZedtimeAnalytics {
	res := ZedtimeAnalytics{}

	if len(wave.Zedtimes) > 0 {
		res.ZedtimeDuration.Min = wave.Zedtimes[0].MetaData.Duration
		res.ZedtimeDuration.Max = wave.Zedtimes[0].MetaData.Duration
	}

	if len(wave.Zedtimes) > 1 {
		res.TimeBetweenZedtimes.Min = float64(wave.Zedtimes[1].TicksSinceLast) / 100
		res.TimeBetweenZedtimes.Max = float64(wave.Zedtimes[1].TicksSinceLast) / 100
	}

	for i := range wave.Zedtimes {
		item := wave.Zedtimes[i]

		res.AvgExtendsCount += float64(item.MetaData.ExtendsCount)
		res.ZedtimeDuration.Avg += item.MetaData.Duration
		res.TimeBetweenZedtimes.Avg += float64(item.TicksSinceLast)

		if item.MetaData.Duration > res.ZedtimeDuration.Max {
			res.ZedtimeDuration.Max = item.MetaData.Duration
		}

		if item.MetaData.Duration < res.ZedtimeDuration.Min {
			res.ZedtimeDuration.Min = item.MetaData.Duration
		}

		if i == 0 {
			continue
		}

		if float64(item.TicksSinceLast)/100 > res.TimeBetweenZedtimes.Max {
			res.TimeBetweenZedtimes.Max = float64(item.TicksSinceLast) / 100
		}

		if float64(item.TicksSinceLast)/100 < res.TimeBetweenZedtimes.Min {
			res.TimeBetweenZedtimes.Min = float64(item.TicksSinceLast) / 100
		}
	}

	if len(wave.Zedtimes) > 0 {
		res.TotalZedtimes = len(wave.Zedtimes)

		res.FirstZedtimeTick.Min = float64(wave.Zedtimes[0].MetaData.StartTick - wave.MetaData.StartTick)
		res.FirstZedtimeTick.Max = res.FirstZedtimeTick.Min
		res.FirstZedtimeTick.Avg = res.FirstZedtimeTick.Min

		res.AvgExtendsCount /= float64(len(wave.Zedtimes))
		res.ZedtimeDuration.Avg /= float64(len(wave.Zedtimes))

		if len(wave.Zedtimes) > 1 {
			res.TimeBetweenZedtimes.Avg /= float64(len(wave.Zedtimes)-1) * 100
		} else {
			res.TimeBetweenZedtimes.Avg = 0
		}

		if res.AvgExtendsCount > 0 {
			res.AvgExtendDuration = (res.ZedtimeDuration.Avg - 3) / res.AvgExtendsCount
		}
	}

	return &res
}

func (wave *DemoRecordAnalysisWave) calcSummary() *Summary {
	res := Summary{
		ZedsKilled: &ZedCounter{},
	}

	if len(wave.ZedsLeft) > 0 {
		res.WaveSize = wave.ZedsLeft[0].ZedsLeft
		res.ZedsLeft = wave.ZedsLeft[len(wave.ZedsLeft)-1].ZedsLeft
		res.CompletionPercent = 1 - float64(res.ZedsLeft)/float64(res.WaveSize)
	}

	res.Duration = float64(wave.MetaData.EndTick-wave.MetaData.StartTick) / 100

	for i := range wave.PlayerEvents.Kills {
		kill := wave.PlayerEvents.Kills[i]

		if kill.IsLarge() {
			res.ZedsKilled.Large += 1
		} else if kill.IsMedium() {
			res.ZedsKilled.Medium += 1
		} else if kill.IsTrash() {
			res.ZedsKilled.Trash += 1
		} else if kill.IsBoss() {
			res.ZedsKilled.Boss += 1
		}

		res.ZedsKilled.Total += 1
	}

	if res.ZedsKilled.Total > 0 {
		res.TrashPercent = float64(res.ZedsKilled.Trash) / float64(res.ZedsKilled.Total)
		res.MediumPercent = float64(res.ZedsKilled.Medium) / float64(res.ZedsKilled.Total)
		res.LargePercent = float64(res.ZedsKilled.Large) / float64(res.ZedsKilled.Total)
		res.AvgKillsPerSecond = float64(res.ZedsKilled.Total) / res.Duration
	}

	return &res
}

func (wave *DemoRecordAnalysisWave) calcBuffsUptime() *BuffsUptimeAnalytics {
	res := BuffsUptimeAnalytics{
		BuffedTicks: 0,
		TotalTicks:  0,

		Detailed: []*DemoRecordAnalysisWaveBuffsUptime{},
	}

	type PlayerBuffs struct {
		Buffs []*DemoRecordParsedEventBuff

		TotalTicks int
		Perk       int
	}

	playerBuffs := map[int]*PlayerBuffs{}
	playerDeathTicks := map[int]int{}
	maxBuffDurationInTicks := 500

	for i := range wave.PlayerEvents.Perks {
		item := wave.PlayerEvents.Perks[i]

		playerBuffs[item.UserId] = &PlayerBuffs{
			Perk: item.Perk,
		}
	}

	for i := range wave.PlayerEvents.Deaths {
		item := wave.PlayerEvents.Deaths[i]

		playerDeathTicks[item.UserId] = item.Tick
	}

	for i := range wave.PlayerEvents.Buffs {
		item := wave.PlayerEvents.Buffs[i]

		deathTick := playerDeathTicks[item.UserId]

		if deathTick > 0 && item.Tick >= deathTick {
			continue
		}

		if data, ok := playerBuffs[item.UserId]; ok {
			data.Buffs = append(data.Buffs, item)
		} else {
			playerBuffs[item.UserId] = &PlayerBuffs{}
			data := playerBuffs[item.UserId]
			data.Buffs = append(data.Buffs, item)
		}
	}

	for userId, deathTick := range playerDeathTicks {
		item := DemoRecordParsedEventBuff{
			UserId:   userId,
			Tick:     deathTick,
			MaxBuffs: -1,
		}

		if data, ok := playerBuffs[userId]; ok {
			data.Buffs = append(data.Buffs, &item)
		} else {
			playerBuffs[userId] = &PlayerBuffs{}
			data := playerBuffs[userId]
			data.Buffs = append(data.Buffs, &item)
		}
	}

	for userId, data := range playerBuffs {
		for buffIdx := range data.Buffs {
			if buffIdx == 0 {
				continue
			}

			prevBuff := data.Buffs[buffIdx-1]
			curBuff := data.Buffs[buffIdx]

			if prevBuff.MaxBuffs > 0 {
				data.TotalTicks += curBuff.Tick - prevBuff.Tick
			}
		}

		if len(data.Buffs) > 0 {
			lastBuff := data.Buffs[len(data.Buffs)-1]
			if lastBuff.MaxBuffs > 0 {
				// Last buff is still active
				data.TotalTicks += min(maxBuffDurationInTicks, wave.MetaData.EndTick-lastBuff.Tick)
			}

			endTick := wave.MetaData.EndTick
			if lastBuff.MaxBuffs < 0 {
				endTick = lastBuff.Tick
			}

			item := DemoRecordAnalysisWaveBuffsUptime{
				UserId:      userId,
				BuffedTicks: data.TotalTicks,
				TotalTicks:  endTick - wave.MetaData.StartTick,
			}

			res.Detailed = append(res.Detailed, &item)

			// Consider all buffs except medic's, because medic's buffs depend on damage taken by team
			if data.Perk != models.Medic {
				res.BuffedTicks += item.BuffedTicks
				res.TotalTicks += item.TotalTicks
			}
		}
	}

	return &res
}
