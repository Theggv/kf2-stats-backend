package demorecord

type DemoRecordAnalysisWaveBuffsUptime struct {
	UserId     int     `json:"user_index"`
	TotalTicks int     `json:"total_ticks"`
	Percent    float64 `json:"percent"`
}

type DemoRecordAnalysisWaveDifficultyItem struct {
	Tick int `json:"tick"`

	Value float64 `json:"value"`
}

type DemoRecordAnalysisZedtime struct {
	MetaData *DemoRecordParsedZedtime `json:"meta_data"`

	TicksSinceLast int `json:"ticks_since_last"`

	TotalKills int `json:"total_kills"`
	LargeKills int `json:"large_kills"`
	HuskKills  int `json:"husk_kills"`
	SirenKills int `json:"siren_kills"`
}

type Metric struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
	Avg float64 `json:"avg"`
}

type ZedCounter struct {
	Total  float64 `json:"total"`
	Trash  float64 `json:"trash"`
	Medium float64 `json:"medium"`
	Large  float64 `json:"large"`
	Boss   float64 `json:"boss"`
}

type ZedtimeAnalytics struct {
	TotalZedtimes int `json:"total_zt_count"`

	FirstZedtimeTick    Metric `json:"first_zt_tick"`
	ZedtimeDuration     Metric `json:"zt_duration_seconds"`
	TimeBetweenZedtimes Metric `json:"time_between_zt_seconds"`

	NonNullZedtimesCount       int `json:"-"`
	NonNullTimeBetweenZedtimes int `json:"-"`

	AvgExtendsCount   float64 `json:"avg_extends_count"`
	AvgExtendDuration float64 `json:"avg_extend_duration"`
}

type Summary struct {
	WaveSize          int     `json:"wave_size"`
	ZedsLeft          int     `json:"zeds_left"`
	CompletionPercent float64 `json:"completion_percent"`

	Duration          float64 `json:"duration"`
	AvgKillsPerSecond float64 `json:"avg_kills_per_second"`

	ZedsKilled *ZedCounter `json:"zeds_killed"`

	TrashPercent  float64 `json:"trash_percent"`
	MediumPercent float64 `json:"medium_percent"`
	LargePercent  float64 `json:"large_percent"`
}

type DemoRecordAnalysisWaveAnalytics struct {
	Summary    *Summary             `json:"summary"`
	Difficulty *DifficultyAnalytics `json:"difficulty"`
	Zedtime    *ZedtimeAnalytics    `json:"zedtime"`

	BuffsUptime []*DemoRecordAnalysisWaveBuffsUptime `json:"buffs_uptime"`
}

type DemoRecordAnalysisWave struct {
	MetaData *DemoRecordParsedWave `json:"meta_data"`

	Analytics *DemoRecordAnalysisWaveAnalytics `json:"analytics"`

	Zedtimes []*DemoRecordAnalysisZedtime     `json:"zedtimes"`
	ZedsLeft []*DemoRecordParsedEventZedsLeft `json:"zeds_left"`

	PlayerEvents *DemoRecordParsedPlayerEvents `json:"player_events"`
}

type DemoRecordAnalysisAnalytics struct {
	Summary    *Summary             `json:"summary"`
	Difficulty *DifficultyAnalytics `json:"difficulty"`
	Zedtime    *ZedtimeAnalytics    `json:"zedtime"`
}

type DemoRecordAnalysis struct {
	Version   byte `json:"protocol_version"`
	SessionId int  `json:"session_id"`

	StartTick int `json:"start_tick"`
	EndTick   int `json:"end_tick"`

	Analytics *DemoRecordAnalysisAnalytics `json:"analytics"`

	Players []*DemoRecordParsedPlayer `json:"players"`
	Waves   []*DemoRecordAnalysisWave `json:"waves"`
}

func (demo *DemoRecordParsed) Analyze() *DemoRecordAnalysis {
	res := DemoRecordAnalysis{
		Version:   demo.Version,
		SessionId: demo.SessionId,
		StartTick: demo.StartTick,
		EndTick:   demo.EndTick,
		Players:   demo.Players,
		Analytics: &DemoRecordAnalysisAnalytics{
			Summary: &Summary{},
			Zedtime: &ZedtimeAnalytics{},
		},
	}

	for i := range demo.WaveEvents.Waves {
		wave := demo.WaveEvents.Waves[i]

		res.Waves = append(res.Waves, demo.analyzeWave(wave))
	}

	// diff := []float64{}
	// for _, wave := range res.Waves {
	// 	for _, tick := range wave.Analytics.Difficulty.Buckets {
	// 		diff = append(diff, tick.Score)
	// 	}
	// }

	// mean, stddev := stat.MeanStdDev(diff, nil)
	// fmt.Printf("mean: %v stddev: %v\n", mean, stddev)

	{
		// Zedtime analytics
		analytics := res.Analytics.Zedtime

		for i := range res.Waves {
			item := res.Waves[i].Analytics.Zedtime

			analytics.TotalZedtimes += item.TotalZedtimes

			analytics.AvgExtendsCount += item.AvgExtendsCount

			if item.TimeBetweenZedtimes.Avg > 0 {
				analytics.NonNullTimeBetweenZedtimes += 1
				analytics.TimeBetweenZedtimes.Avg += item.TimeBetweenZedtimes.Avg
			}

			if item.FirstZedtimeTick.Avg > 0 {
				analytics.NonNullZedtimesCount += 1
				analytics.FirstZedtimeTick.Avg += item.FirstZedtimeTick.Avg
				analytics.ZedtimeDuration.Avg += item.ZedtimeDuration.Avg
			}

			if item.FirstZedtimeTick.Max > analytics.FirstZedtimeTick.Max {
				analytics.FirstZedtimeTick.Max = item.FirstZedtimeTick.Max
			}

			if item.FirstZedtimeTick.Min > 0 &&
				(item.FirstZedtimeTick.Min < analytics.FirstZedtimeTick.Min ||
					analytics.FirstZedtimeTick.Min == 0) {
				analytics.FirstZedtimeTick.Min = item.FirstZedtimeTick.Min
			}

			if item.ZedtimeDuration.Max > analytics.ZedtimeDuration.Max {
				analytics.ZedtimeDuration.Max = item.ZedtimeDuration.Max
			}

			if item.ZedtimeDuration.Min > 0 &&
				(item.ZedtimeDuration.Min < analytics.ZedtimeDuration.Min ||
					analytics.ZedtimeDuration.Min == 0) {
				analytics.ZedtimeDuration.Min = item.ZedtimeDuration.Min
			}

			if item.TimeBetweenZedtimes.Max > analytics.TimeBetweenZedtimes.Max {
				analytics.TimeBetweenZedtimes.Max = item.TimeBetweenZedtimes.Max
			}

			if item.TimeBetweenZedtimes.Min > 0 &&
				(item.TimeBetweenZedtimes.Min < analytics.TimeBetweenZedtimes.Min ||
					analytics.TimeBetweenZedtimes.Min == 0) {
				analytics.TimeBetweenZedtimes.Min = item.TimeBetweenZedtimes.Min
			}
		}

		if analytics.TotalZedtimes > 0 {
			if analytics.NonNullZedtimesCount > 0 {
				analytics.AvgExtendsCount /= float64(analytics.NonNullZedtimesCount)
				analytics.FirstZedtimeTick.Avg /= float64(analytics.NonNullZedtimesCount)
				analytics.ZedtimeDuration.Avg /= float64(analytics.NonNullZedtimesCount)
			}

			if analytics.NonNullTimeBetweenZedtimes > 0 {
				analytics.TimeBetweenZedtimes.Avg /= float64(analytics.NonNullTimeBetweenZedtimes)
			}

			if analytics.AvgExtendsCount > 0 {
				analytics.AvgExtendDuration = (analytics.ZedtimeDuration.Avg - 3) / analytics.AvgExtendsCount
			}
		}
	}

	return &res
}

func (demo *DemoRecordParsed) analyzeWave(wave *DemoRecordParsedWave) *DemoRecordAnalysisWave {
	res := DemoRecordAnalysisWave{
		MetaData: wave,
		Analytics: &DemoRecordAnalysisWaveAnalytics{
			Summary:     &Summary{},
			Difficulty:  &DifficultyAnalytics{},
			Zedtime:     &ZedtimeAnalytics{},
			BuffsUptime: []*DemoRecordAnalysisWaveBuffsUptime{},
		},
		Zedtimes:     []*DemoRecordAnalysisZedtime{},
		ZedsLeft:     []*DemoRecordParsedEventZedsLeft{},
		PlayerEvents: demo.analyzePlayerEvents(wave),
	}

	zedTimes := filterByRange(
		demo.WaveEvents.ZedTimes,
		func(item *DemoRecordParsedZedtime) int {
			return item.StartTick
		}, wave.StartTick, wave.EndTick,
	)

	res.Zedtimes = append(res.Zedtimes, demo.analyzeWaveZedtimes(zedTimes)...)

	zedsLeft := filterByRange(
		demo.WaveEvents.ZedsLeft,
		func(item *DemoRecordParsedEventZedsLeft) int {
			return item.Tick
		}, wave.StartTick, wave.EndTick,
	)

	res.ZedsLeft = append(res.ZedsLeft, zedsLeft...)

	{
		// Zedtime wave analytics
		analytics := res.Analytics.Zedtime

		if len(res.Zedtimes) > 0 {
			analytics.ZedtimeDuration.Min = res.Zedtimes[0].MetaData.Duration
			analytics.ZedtimeDuration.Max = res.Zedtimes[0].MetaData.Duration
		}

		if len(res.Zedtimes) > 1 {
			analytics.TimeBetweenZedtimes.Min = float64(res.Zedtimes[1].TicksSinceLast) / 100
			analytics.TimeBetweenZedtimes.Max = float64(res.Zedtimes[1].TicksSinceLast) / 100
		}

		for i := range res.Zedtimes {
			item := res.Zedtimes[i]

			analytics.AvgExtendsCount += float64(item.MetaData.ExtendsCount)
			analytics.ZedtimeDuration.Avg += item.MetaData.Duration
			analytics.TimeBetweenZedtimes.Avg += float64(item.TicksSinceLast)

			if item.MetaData.Duration > analytics.ZedtimeDuration.Max {
				analytics.ZedtimeDuration.Max = item.MetaData.Duration
			}

			if item.MetaData.Duration < analytics.ZedtimeDuration.Min {
				analytics.ZedtimeDuration.Min = item.MetaData.Duration
			}

			if i == 0 {
				continue
			}

			if float64(item.TicksSinceLast)/100 > analytics.TimeBetweenZedtimes.Max {
				analytics.TimeBetweenZedtimes.Max = float64(item.TicksSinceLast) / 100
			}

			if float64(item.TicksSinceLast)/100 < analytics.TimeBetweenZedtimes.Min {
				analytics.TimeBetweenZedtimes.Min = float64(item.TicksSinceLast) / 100
			}
		}

		if len(res.Zedtimes) > 0 {
			analytics.TotalZedtimes = len(res.Zedtimes)

			analytics.FirstZedtimeTick.Min = float64(res.Zedtimes[0].MetaData.StartTick - res.MetaData.StartTick)
			analytics.FirstZedtimeTick.Max = analytics.FirstZedtimeTick.Min
			analytics.FirstZedtimeTick.Avg = analytics.FirstZedtimeTick.Min

			analytics.AvgExtendsCount /= float64(len(res.Zedtimes))
			analytics.ZedtimeDuration.Avg /= float64(len(res.Zedtimes))

			if len(res.Zedtimes) > 1 {
				analytics.TimeBetweenZedtimes.Avg /= float64(len(res.Zedtimes)-1) * 100
			} else {
				analytics.TimeBetweenZedtimes.Avg = 0
			}

			if analytics.AvgExtendsCount > 0 {
				analytics.AvgExtendDuration = (analytics.ZedtimeDuration.Avg - 3) / analytics.AvgExtendsCount
			}
		}
	}

	res.Analytics.BuffsUptime = calcWaveBuffsUptime(&res)

	res.generateSummary()
	res.calcDifficulty(100, 500)

	return &res
}

func (demo *DemoRecordParsed) analyzePlayerEvents(
	wave *DemoRecordParsedWave,
) *DemoRecordParsedPlayerEvents {
	res := DemoRecordParsedPlayerEvents{
		ConnectionLog: []*DemoRecordParsedEventConnection{},
		Perks:         []*DemoRecordParsedEventPerkChange{},
		Kills:         []*DemoRecordParsedEventKill{},
		Buffs:         []*DemoRecordParsedEventBuff{},
		Deaths:        []*DemoRecordParsedEventDeath{},
		HuskRages:     []*DemoRecordParsedEventHuskRage{},
		HealthChanges: []*DemoRecordParsedEventHpChange{},
	}

	for i := range demo.PlayerEvents.Perks {
		userId := demo.PlayerEvents.Perks[i].UserId

		if lastHpChange := findLastLower(
			filter(
				demo.PlayerEvents.HealthChanges,
				func(item *DemoRecordParsedEventHpChange) int {
					return item.UserId
				}, userId,
			),
			func(item *DemoRecordParsedEventHpChange) int {
				return item.Tick
			},
			wave.StartTick,
		); lastHpChange != nil {
			res.HealthChanges = append(res.HealthChanges, *lastHpChange)
		}

		res.Buffs = append(res.Buffs, &DemoRecordParsedEventBuff{
			Tick:     wave.StartTick,
			UserId:   userId,
			MaxBuffs: 0,
		})
	}

	res.Perks = append(res.Perks,
		filterByRange(demo.PlayerEvents.Perks, func(item *DemoRecordParsedEventPerkChange) int {
			return item.Tick
		}, wave.StartTick, wave.EndTick)...,
	)

	res.ConnectionLog = append(res.ConnectionLog,
		filterByRange(demo.PlayerEvents.ConnectionLog, func(item *DemoRecordParsedEventConnection) int {
			return item.Tick
		}, wave.StartTick, wave.EndTick)...,
	)

	res.Deaths = append(res.Deaths,
		filterByRange(demo.PlayerEvents.Deaths, func(item *DemoRecordParsedEventDeath) int {
			return item.Tick
		}, wave.StartTick, wave.EndTick)...,
	)

	res.Kills = append(res.Kills,
		filterByRange(demo.PlayerEvents.Kills, func(item *DemoRecordParsedEventKill) int {
			return item.Tick
		}, wave.StartTick, wave.EndTick)...,
	)

	res.HuskRages = append(res.HuskRages,
		filterByRange(demo.PlayerEvents.HuskRages, func(item *DemoRecordParsedEventHuskRage) int {
			return item.Tick
		}, wave.StartTick, wave.EndTick)...,
	)

	res.HealthChanges = append(res.HealthChanges,
		filterByRange(demo.PlayerEvents.HealthChanges, func(item *DemoRecordParsedEventHpChange) int {
			return item.Tick
		}, wave.StartTick, wave.EndTick)...,
	)

	res.Buffs = append(res.Buffs,
		filterByRange(demo.PlayerEvents.Buffs, func(item *DemoRecordParsedEventBuff) int {
			return item.Tick
		}, wave.StartTick, wave.EndTick)...,
	)

	return &res
}

func (demo *DemoRecordParsed) analyzeWaveZedtimes(
	zedTimes []*DemoRecordParsedZedtime,
) []*DemoRecordAnalysisZedtime {
	items := []*DemoRecordAnalysisZedtime{}

	for i := range zedTimes {
		item := DemoRecordAnalysisZedtime{
			MetaData: zedTimes[i],
		}

		if i > 0 {
			item.TicksSinceLast = zedTimes[i].StartTick - zedTimes[i-1].EndTick
		}

		kills := filterByRange(demo.PlayerEvents.Kills, func(item *DemoRecordParsedEventKill) int {
			return item.Tick
		}, item.MetaData.StartTick, item.MetaData.EndTick)

		for killIdx := range kills {
			kill := kills[killIdx]

			if kill.IsLarge() {
				item.LargeKills += 1
			} else if kill.IsHusk() {
				item.HuskKills += 1
			} else if kill.IsSiren() {
				item.SirenKills += 1
			}

			item.TotalKills += 1
		}

		items = append(items, &item)
	}

	return items
}

func (wave *DemoRecordAnalysisWave) generateSummary() {
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

	wave.Analytics.Summary = &res
}

func calcWaveBuffsUptime(wave *DemoRecordAnalysisWave) []*DemoRecordAnalysisWaveBuffsUptime {
	res := []*DemoRecordAnalysisWaveBuffsUptime{}

	type PlayerBuffs struct {
		Buffs []*DemoRecordParsedEventBuff

		TotalTicks int
	}

	playerBuffs := map[int]*PlayerBuffs{}
	maxBuffDurationInTicks := 500

	for i := range wave.PlayerEvents.Perks {
		item := wave.PlayerEvents.Perks[i]

		playerBuffs[item.UserId] = &PlayerBuffs{}
	}

	for i := range wave.PlayerEvents.Buffs {
		item := wave.PlayerEvents.Buffs[i]

		if data, ok := playerBuffs[item.UserId]; ok {
			data.Buffs = append(data.Buffs, item)
		} else {
			playerBuffs[item.UserId] = &PlayerBuffs{}
			data := playerBuffs[item.UserId]
			data.Buffs = append(data.Buffs, item)
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
		}

		res = append(res, &DemoRecordAnalysisWaveBuffsUptime{
			UserId:     userId,
			TotalTicks: data.TotalTicks,
			Percent:    float64(data.TotalTicks) / float64(wave.MetaData.EndTick-wave.MetaData.StartTick),
		})
	}

	return res
}
