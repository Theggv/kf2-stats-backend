package demorecord

import (
	"errors"
	"fmt"
)

type DemoRecordAnalysisPlayer struct {
	UserId   int    `json:"user_id"`
	UserType int    `json:"user_type"`
	UniqueId string `json:"unique_id"`
}

type DemoRecordAnalysisWavePerk struct {
	UserId int `json:"user_id"`
	Perk   int `json:"perk"`
}

type DemoRecordAnalysisWaveZedtime struct {
	StartTick int   `json:"start_tick"`
	EndTick   int   `json:"end_tick"`
	Ticks     []int `json:"ticks"`

	Duration     float32 `json:"duration"`
	ExtendsCount int     `json:"extends_count"`

	TotalKills int `json:"total_kills"`
	LargeKills int `json:"large_kills"`
	HuskKills  int `json:"husk_kills"`
	SirenKills int `json:"siren_kills"`
}

type DemoRecordAnalysisWaveKill struct {
	Tick int `json:"tick"`

	UserId int `json:"user_id"`
	Zed    int `json:"zed"`
}

type DemoRecordAnalysisWaveBuff struct {
	Tick int `json:"tick"`

	UserId   int `json:"user_id"`
	MaxBuffs int `json:"max_buffs"`
}

type DemoRecordAnalysisWaveHpChange struct {
	Tick int `json:"tick"`

	UserId int `json:"user_id"`
	Health int `json:"health"`
	Armor  int `json:"armor"`
}

type DemoRecordAnalysisWaveBuffsUptime struct {
	UserId     int     `json:"user_id"`
	TotalTicks int     `json:"total_ticks"`
	Percent    float64 `json:"percent"`
}

type DemoRecordAnalysisWaveDifficultyItem struct {
	Tick int `json:"tick"`

	TrashKills  int `json:"trash_kills"`
	MediumKills int `json:"medium_kills"`
	LargeKills  int `json:"large_kills"`
}

type DemoRecordAnalysisWaveZedsLeft struct {
	Tick int `json:"tick"`

	ZedsLeft int `json:"zeds_left"`
}

type DemoRecordAnalysisWaveDifficulty struct {
	Step        int `json:"step"`
	PeriodTicks int `json:"period"`

	Ticks []*DemoRecordAnalysisWaveDifficultyItem `json:"ticks"`
}

type DemoRecordAnalysisWave struct {
	Wave int `json:"wave"`

	StartTick int `json:"start_tick"`
	EndTick   int `json:"end_tick"`

	RawEvents []*DemoRecordRawEvent            `json:"raw_events"`
	Perks     []*DemoRecordAnalysisWavePerk    `json:"perks"`
	ZedTimes  []*DemoRecordAnalysisWaveZedtime `json:"zed_times,omitempty"`
	Kills     []*DemoRecordAnalysisWaveKill    `json:"kills"`
	Buffs     []*DemoRecordAnalysisWaveBuff    `json:"buffs"`

	ZedsLeft []*DemoRecordAnalysisWaveZedsLeft `json:"zeds_left"`

	HealthChanges []*DemoRecordAnalysisWaveHpChange `json:"hp_changes"`

	BuffsUptime []*DemoRecordAnalysisWaveBuffsUptime `json:"buffs_uptime"`
	Difficulty  *DemoRecordAnalysisWaveDifficulty    `json:"difficulty"`
}

type DemoRecordAnalysisConnection struct {
	Tick int `json:"tick"`

	UserId int `json:"user_id"`
	Type   int `json:"type"`
}

type DemoRecordAnalysis struct {
	Header *DemoRecordHeader `json:"header"`

	StartTick int `json:"start_tick"`
	EndTick   int `json:"end_tick"`

	Players []*DemoRecordAnalysisPlayer `json:"players"`
	Waves   []*DemoRecordAnalysisWave   `json:"waves"`

	Connections []*DemoRecordAnalysisConnection `json:"connection"`
}

func Transform(demo *DemoRecord) (*DemoRecordAnalysis, error) {
	analysis := DemoRecordAnalysis{
		Header: demo.Header,
	}

	if len(demo.Events) > 0 {
		analysis.StartTick = demo.Events[0].Tick
		analysis.EndTick = demo.Events[len(demo.Events)-1].Tick
	}

	{
		joinEvents := filterEventsByType(demo.Events, byte(PlayerJoin))

		unique := map[int]*DemoRecordAnalysisPlayer{}

		for i := range joinEvents {
			event := joinEvents[i]

			player := DemoRecordAnalysisPlayer{
				UserId:   int(event.Data["user_id"].(byte)),
				UserType: int(event.Data["user_type"].(byte)),
				UniqueId: event.Data["unique_id"].(string),
			}

			unique[player.UserId] = &player
		}

		for _, item := range unique {
			analysis.Players = append(analysis.Players, item)
		}
	}

	{
		connectionEvents := filterEventsByType(demo.Events, byte(PlayerJoin), byte(PlayerDisconnect))

		for i := range connectionEvents {
			event := connectionEvents[i]

			analysis.Connections = append(analysis.Connections, &DemoRecordAnalysisConnection{
				Tick:   event.Tick,
				UserId: int(event.Data["user_id"].(byte)),
				Type:   int(event.Type),
			})
		}
	}

	{
		waveStart := 0
		waveEnd := 0
		for i := range demo.Events {
			event := demo.Events[i]

			if event.Type == byte(GlobalWaveStart) {
				waveStart = i
			} else if event.Type == byte(GlobalWaveEnd) {
				waveEnd = i

				wave, err := transformWave(demo.Events[waveStart : waveEnd+1])
				if err != nil {
					return nil, err
				}

				analysis.Waves = append(analysis.Waves, wave)

				if waveEnd < waveStart {
					return nil, errors.New(fmt.Sprintf("waveEnd < waveStart at pos %v", waveEnd))
				}
			}
		}

		// Detect if last wave is not finished
		if waveEnd < waveStart {
			wave, err := transformWave(demo.Events[waveStart:])
			if err != nil {
				return nil, err
			}

			analysis.Waves = append(analysis.Waves, wave)
		}
	}

	for i := range analysis.Waves {
		wave := analysis.Waves[i]

		calcWaveBuffsUptime(wave)
	}

	return &analysis, nil
}

func transformWave(events []*DemoRecordRawEvent) (*DemoRecordAnalysisWave, error) {
	wave := DemoRecordAnalysisWave{
		Wave:      int(events[0].Data["wave"].(byte)),
		RawEvents: events,
		StartTick: events[0].Tick,
		EndTick:   events[len(events)-1].Tick,
	}

	{
		perkEvents := filterEventsByType(events, byte(PlayerPerk))

		for i := range perkEvents {
			event := perkEvents[i]

			wave.Perks = append(wave.Perks, &DemoRecordAnalysisWavePerk{
				UserId: int(event.Data["user_id"].(byte)),
				Perk:   int(event.Data["perk"].(byte)),
			})
		}
	}

	{
		zedtimeEvents := filterEventsByType(events, byte(GlobalZedTime))
		zedtimeDividers := []int{}

		for i := range zedtimeEvents {
			if i == 0 {
				continue
			}

			if zedtimeEvents[i].Tick-zedtimeEvents[i-1].Tick > 300 {
				zedtimeDividers = append(zedtimeDividers, i-1)
			}
		}

		if len(zedtimeEvents) > 0 {
			zedtimeDividers = append(zedtimeDividers, len(zedtimeEvents)-1)
		}

		for i, rangeEndIndex := range zedtimeDividers {
			rangeStartIndex := 0
			if i > 0 {
				rangeStartIndex = zedtimeDividers[i-1] + 1
			}

			item := DemoRecordAnalysisWaveZedtime{
				Duration:     float32(zedtimeEvents[rangeEndIndex].Tick+300-zedtimeEvents[rangeStartIndex].Tick) / 100,
				ExtendsCount: rangeEndIndex - rangeStartIndex,
			}

			for j := rangeStartIndex; j <= rangeEndIndex; j++ {
				item.Ticks = append(item.Ticks, zedtimeEvents[j].Tick)
			}

			item.Ticks = append(item.Ticks, zedtimeEvents[rangeEndIndex].Tick+300)
			item.StartTick = item.Ticks[0]
			item.EndTick = item.Ticks[len(item.Ticks)-1]

			wave.ZedTimes = append(wave.ZedTimes, &item)
		}
	}

	{
		killEvents := filterEventsByType(events, byte(EventKill))

		for i := range killEvents {
			event := killEvents[i]

			wave.Kills = append(wave.Kills, &DemoRecordAnalysisWaveKill{
				Tick:   event.Tick,
				UserId: int(event.Data["user_id"].(byte)),
				Zed:    int(event.Data["zed"].(byte)),
			})
		}
	}

	{
		buffEvents := filterEventsByType(events, byte(EventBuffs), byte(PlayerDied))

		for i := range buffEvents {
			event := buffEvents[i]

			if event.Type == byte(EventBuffs) {
				wave.Buffs = append(wave.Buffs, &DemoRecordAnalysisWaveBuff{
					Tick:     event.Tick,
					UserId:   int(event.Data["user_id"].(byte)),
					MaxBuffs: int(event.Data["max_buffs"].(byte)),
				})
			} else {
				wave.Buffs = append(wave.Buffs, &DemoRecordAnalysisWaveBuff{
					Tick:     event.Tick,
					UserId:   int(event.Data["user_id"].(byte)),
					MaxBuffs: 0,
				})
			}
		}
	}

	{
		hpChangeEvents := filterEventsByType(events, byte(EventHpChange), byte(PlayerDied))

		for i := range hpChangeEvents {
			event := hpChangeEvents[i]

			if event.Type == byte(EventHpChange) {
				wave.HealthChanges = append(wave.HealthChanges, &DemoRecordAnalysisWaveHpChange{
					Tick:   event.Tick,
					UserId: int(event.Data["user_id"].(byte)),
					Health: event.Data["health"].(int),
					Armor:  int(event.Data["armor"].(byte)),
				})
			} else {
				wave.HealthChanges = append(wave.HealthChanges, &DemoRecordAnalysisWaveHpChange{
					Tick:   event.Tick,
					UserId: int(event.Data["user_id"].(byte)),
					Health: 0,
					Armor:  0,
				})
			}
		}
	}

	{
		zedsLeftEvents := filterEventsByType(events, byte(GlobalZedsLeft))

		for i := range zedsLeftEvents {
			event := zedsLeftEvents[i]

			wave.ZedsLeft = append(wave.ZedsLeft, &DemoRecordAnalysisWaveZedsLeft{
				Tick:     event.Tick,
				ZedsLeft: event.Data["zeds_left"].(int),
			})
		}
	}

	analyzeWaveZedtimes(&wave)
	wave.Difficulty = analyzeWaveDifficulty(&wave)

	return &wave, nil
}

func analyzeWaveZedtimes(wave *DemoRecordAnalysisWave) {
	lastIdx := 0

	for i := range wave.ZedTimes {
		zedtime := wave.ZedTimes[i]

		for j := lastIdx; j < len(wave.Kills); j++ {
			kill := wave.Kills[j]

			if kill.Tick < zedtime.Ticks[0] {
				continue
			} else if kill.Tick > zedtime.Ticks[len(zedtime.Ticks)-1]+300 {
				lastIdx = j
				break
			} else {
				if kill.Zed == 7 || kill.Zed == 8 || kill.Zed == 9 {
					zedtime.LargeKills += 1
				} else {
					if kill.Tick > zedtime.Ticks[len(zedtime.Ticks)-1] {
						continue
					}

					if kill.Zed == 11 {
						zedtime.SirenKills += 1
					} else if kill.Zed == 12 {
						zedtime.HuskKills += 1
					}

					zedtime.TotalKills += 1
				}
			}
		}
	}
}

func analyzeWaveDifficulty(wave *DemoRecordAnalysisWave) *DemoRecordAnalysisWaveDifficulty {
	difficulty := DemoRecordAnalysisWaveDifficulty{
		Step:        100,
		PeriodTicks: 3000,
	}

	startIdx := 0
	endIdx := -1
	tick := DemoRecordAnalysisWaveDifficultyItem{}

	updateKills := func(kill *DemoRecordAnalysisWaveKill, value int) {
		if kill.Zed == 7 || kill.Zed == 8 || kill.Zed == 9 {
			tick.LargeKills += value
		} else if kill.Zed == 10 || kill.Zed == 11 || kill.Zed == 12 {
			tick.MediumKills += value
		} else {
			tick.TrashKills += value
		}
	}

	for tickOffset := 0; tickOffset < wave.EndTick-wave.StartTick+difficulty.Step; tickOffset += difficulty.Step {
		if tickOffset >= difficulty.PeriodTicks {
			for i := startIdx; i < len(wave.Kills); i++ {
				kill := wave.Kills[i]

				if kill.Tick > wave.StartTick+tickOffset-difficulty.PeriodTicks {
					break
				}

				startIdx = i + 1
				updateKills(kill, -1)
			}
		}

		for i := endIdx + 1; i < len(wave.Kills); i++ {
			kill := wave.Kills[i]

			if kill.Tick > wave.StartTick+tickOffset {
				break
			}

			endIdx = i
			updateKills(kill, 1)
		}

		difficulty.Ticks = append(difficulty.Ticks, &DemoRecordAnalysisWaveDifficultyItem{
			Tick:        tickOffset,
			TrashKills:  tick.TrashKills,
			MediumKills: tick.MediumKills,
			LargeKills:  tick.LargeKills,
		})
	}

	return &difficulty
}

func calcWaveBuffsUptime(wave *DemoRecordAnalysisWave) {
	type PlayerBuffs struct {
		Buffs []*DemoRecordAnalysisWaveBuff

		TotalTicks int
	}

	playerBuffs := map[int]*PlayerBuffs{}
	maxBuffDurationInTicks := 500

	for i := range wave.Perks {
		item := wave.Perks[i]

		playerBuffs[item.UserId] = &PlayerBuffs{}
	}

	for i := range wave.Buffs {
		item := wave.Buffs[i]

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
				data.TotalTicks += min(maxBuffDurationInTicks, wave.EndTick-lastBuff.Tick)
			}
		}

		wave.BuffsUptime = append(wave.BuffsUptime, &DemoRecordAnalysisWaveBuffsUptime{
			UserId:     userId,
			TotalTicks: data.TotalTicks,
			Percent:    float64(data.TotalTicks) / float64(wave.EndTick-wave.StartTick),
		})
	}
}

func filterEventsByType(events []*DemoRecordRawEvent, eventTypes ...byte) []*DemoRecordRawEvent {
	filtered := []*DemoRecordRawEvent{}

	for i := range events {
		for j := range eventTypes {
			if events[i].Type == eventTypes[j] {
				filtered = append(filtered, events[i])
			}
		}
	}

	return filtered
}
