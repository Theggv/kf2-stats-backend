package demorecord

import "github.com/theggv/kf2-stats-backend/pkg/common/models"

type DemoRecordAnalysisWaveBuffsUptime struct {
	UserId int `json:"user_index"`

	BuffedTicks int `json:"buffed_ticks"`
	TotalTicks  int `json:"total_ticks"`
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

type BuffsUptimeAnalytics struct {
	BuffedTicks int `json:"buffed_ticks"`
	TotalTicks  int `json:"total_ticks"`

	Detailed []*DemoRecordAnalysisWaveBuffsUptime `json:"detailed"`
}

type DemoRecordAnalysisWaveAnalytics struct {
	Summary    *Summary             `json:"summary"`
	Difficulty *DifficultyAnalytics `json:"difficulty"`
	Zedtime    *ZedtimeAnalytics    `json:"zedtime"`

	BuffsUptime *BuffsUptimeAnalytics `json:"buffs_uptime"`
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

	BuffsUptime *BuffsUptimeAnalytics `json:"buffs_uptime"`
}

type DemoRecordPlayers []*DemoRecordParsedPlayer

func (p DemoRecordPlayers) GetByIndex(userIndex int) *models.UserProfile {
	for _, item := range p {
		if item.UserId == userIndex {
			return item.Profile
		}
	}

	return nil
}

type DemoRecordAnalysis struct {
	Version   byte `json:"protocol_version"`
	SessionId int  `json:"session_id"`

	StartTick int `json:"start_tick"`
	EndTick   int `json:"end_tick"`

	Analytics *DemoRecordAnalysisAnalytics `json:"analytics"`

	Players DemoRecordPlayers         `json:"players"`
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
			Summary:     &Summary{},
			Zedtime:     &ZedtimeAnalytics{},
			BuffsUptime: &BuffsUptimeAnalytics{},
		},
	}

	for i := range demo.WaveEvents.Waves {
		wave := demo.WaveEvents.Waves[i]

		res.Waves = append(res.Waves, demo.analyzeWave(wave))
	}

	res.Analytics.Zedtime = res.calcZedtimeAnalytics()
	res.Analytics.BuffsUptime = res.calcBuffsUptime()

	return &res
}

func (demo *DemoRecordParsed) analyzeWave(wave *DemoRecordParsedWave) *DemoRecordAnalysisWave {
	res := DemoRecordAnalysisWave{
		MetaData: wave,
		Analytics: &DemoRecordAnalysisWaveAnalytics{
			Summary:     &Summary{},
			Difficulty:  &DifficultyAnalytics{},
			Zedtime:     &ZedtimeAnalytics{},
			BuffsUptime: &BuffsUptimeAnalytics{},
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

	res.Analytics.Zedtime = res.calcZedtimeAnalytics()
	res.Analytics.Summary = res.calcSummary()

	res.Analytics.BuffsUptime = res.calcBuffsUptime()
	res.Analytics.Difficulty = res.calcDifficulty(100, 1000)

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
		} else {
			res.HealthChanges = append(res.HealthChanges, &DemoRecordParsedEventHpChange{
				Tick:   wave.StartTick,
				UserId: userId,
				Health: 100,
			})
		}

		if buff := findLastLower(
			filter(
				demo.PlayerEvents.Buffs,
				func(item *DemoRecordParsedEventBuff) int {
					return item.UserId
				}, userId,
			),
			func(item *DemoRecordParsedEventBuff) int {
				return item.Tick
			},
			wave.StartTick,
		); buff != nil {
			res.Buffs = append(res.Buffs, &DemoRecordParsedEventBuff{
				Tick:     wave.StartTick,
				UserId:   userId,
				MaxBuffs: (*buff).MaxBuffs,
			})
		} else {
			res.Buffs = append(res.Buffs, &DemoRecordParsedEventBuff{
				Tick:     wave.StartTick,
				UserId:   userId,
				MaxBuffs: 0,
			})
		}
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

func (demo *DemoRecordAnalysis) calcZedtimeAnalytics() *ZedtimeAnalytics {
	res := ZedtimeAnalytics{}

	for i := range demo.Waves {
		item := demo.Waves[i].Analytics.Zedtime

		res.TotalZedtimes += item.TotalZedtimes

		res.AvgExtendsCount += item.AvgExtendsCount

		if item.TimeBetweenZedtimes.Avg > 0 {
			res.NonNullTimeBetweenZedtimes += 1
			res.TimeBetweenZedtimes.Avg += item.TimeBetweenZedtimes.Avg
		}

		if item.FirstZedtimeTick.Avg > 0 {
			res.NonNullZedtimesCount += 1
			res.FirstZedtimeTick.Avg += item.FirstZedtimeTick.Avg
			res.ZedtimeDuration.Avg += item.ZedtimeDuration.Avg
		}

		if item.FirstZedtimeTick.Max > res.FirstZedtimeTick.Max {
			res.FirstZedtimeTick.Max = item.FirstZedtimeTick.Max
		}

		if item.FirstZedtimeTick.Min > 0 &&
			(item.FirstZedtimeTick.Min < res.FirstZedtimeTick.Min ||
				res.FirstZedtimeTick.Min == 0) {
			res.FirstZedtimeTick.Min = item.FirstZedtimeTick.Min
		}

		if item.ZedtimeDuration.Max > res.ZedtimeDuration.Max {
			res.ZedtimeDuration.Max = item.ZedtimeDuration.Max
		}

		if item.ZedtimeDuration.Min > 0 &&
			(item.ZedtimeDuration.Min < res.ZedtimeDuration.Min ||
				res.ZedtimeDuration.Min == 0) {
			res.ZedtimeDuration.Min = item.ZedtimeDuration.Min
		}

		if item.TimeBetweenZedtimes.Max > res.TimeBetweenZedtimes.Max {
			res.TimeBetweenZedtimes.Max = item.TimeBetweenZedtimes.Max
		}

		if item.TimeBetweenZedtimes.Min > 0 &&
			(item.TimeBetweenZedtimes.Min < res.TimeBetweenZedtimes.Min ||
				res.TimeBetweenZedtimes.Min == 0) {
			res.TimeBetweenZedtimes.Min = item.TimeBetweenZedtimes.Min
		}
	}

	if res.TotalZedtimes > 0 {
		if res.NonNullZedtimesCount > 0 {
			res.AvgExtendsCount /= float64(res.NonNullZedtimesCount)
			res.FirstZedtimeTick.Avg /= float64(res.NonNullZedtimesCount)
			res.ZedtimeDuration.Avg /= float64(res.NonNullZedtimesCount)
		}

		if res.NonNullTimeBetweenZedtimes > 0 {
			res.TimeBetweenZedtimes.Avg /= float64(res.NonNullTimeBetweenZedtimes)
		}

		if res.AvgExtendsCount > 0 {
			res.AvgExtendDuration = (res.ZedtimeDuration.Avg - 3) / res.AvgExtendsCount
		}
	}

	return &res
}

func (demo *DemoRecordAnalysis) calcBuffsUptime() *BuffsUptimeAnalytics {
	res := BuffsUptimeAnalytics{
		BuffedTicks: 0,
		TotalTicks:  0,

		Detailed: []*DemoRecordAnalysisWaveBuffsUptime{},
	}

	playerBuffs := map[int]*DemoRecordAnalysisWaveBuffsUptime{}

	for i := range demo.Waves {
		wave := demo.Waves[i]

		data := wave.Analytics.BuffsUptime

		for j := range data.Detailed {
			item := data.Detailed[j]

			if data, ok := playerBuffs[item.UserId]; ok {
				data.BuffedTicks += item.BuffedTicks
				data.TotalTicks += item.TotalTicks
			} else {
				playerBuffs[item.UserId] = &DemoRecordAnalysisWaveBuffsUptime{
					UserId:      item.UserId,
					BuffedTicks: item.BuffedTicks,
					TotalTicks:  item.TotalTicks,
				}
			}
		}

		res.BuffedTicks += data.BuffedTicks
		res.TotalTicks += data.TotalTicks
	}

	for _, data := range playerBuffs {
		res.Detailed = append(res.Detailed, data)
	}

	return &res
}
