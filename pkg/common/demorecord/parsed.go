package demorecord

import (
	"fmt"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

type DemoRecordParsedPlayer struct {
	UserId int `json:"user_index"`

	UniqueId string `json:"auth_id"`
	UserType int    `json:"auth_type"`

	Profile *models.UserProfile `json:"profile,omitempty"`
}

type DemoRecordParsedEventPerkChange struct {
	Tick int `json:"tick"`

	UserId int `json:"user_index"`
	Perk   int `json:"perk"`
}

type DemoRecordParsedEventDeath struct {
	Tick int `json:"tick"`

	UserId int `json:"user_index"`
	Cause  int `json:"cause"`
}

type DemoRecordParsedZedtime struct {
	StartTick int   `json:"start_tick"`
	EndTick   int   `json:"end_tick"`
	Ticks     []int `json:"ticks"`

	Duration     float64 `json:"duration"`
	ExtendsCount int     `json:"extends_count"`
}

type DemoRecordParsedEventKill struct {
	Tick int `json:"tick"`

	UserId int `json:"user_index"`
	Zed    int `json:"zed"`
}

func (data *DemoRecordParsedEventKill) IsScrake() bool {
	return data.Zed == 7
}

func (data *DemoRecordParsedEventKill) IsFleshpound() bool {
	return data.Zed == 8
}

func (data *DemoRecordParsedEventKill) IsMiniFleshpound() bool {
	return data.Zed == 9
}

func (data *DemoRecordParsedEventKill) IsBloat() bool {
	return data.Zed == 10
}

func (data *DemoRecordParsedEventKill) IsSiren() bool {
	return data.Zed == 11
}

func (data *DemoRecordParsedEventKill) IsHusk() bool {
	return data.Zed == 12
}

func (data *DemoRecordParsedEventKill) IsLarge() bool {
	return data.Zed >= 7 && data.Zed <= 9
}

func (data *DemoRecordParsedEventKill) IsMedium() bool {
	return data.Zed >= 10 && data.Zed <= 12 || data.Zed >= 16 && data.Zed <= 18
}

func (data *DemoRecordParsedEventKill) IsBoss() bool {
	return data.Zed >= 19 && data.Zed <= 23
}

func (data *DemoRecordParsedEventKill) IsTrash() bool {
	return data.Zed >= 0 && data.Zed <= 6 || data.Zed >= 13 && data.Zed <= 15
}

type DemoRecordParsedEventBuff struct {
	Tick int `json:"tick"`

	UserId   int `json:"user_index"`
	MaxBuffs int `json:"max_buffs"`
}

type DemoRecordParsedEventHpChange struct {
	Tick int `json:"tick"`

	UserId int `json:"user_index"`
	Health int `json:"health"`
	Armor  int `json:"armor"`
}

type DemoRecordParsedEventConnection struct {
	Tick int `json:"tick"`

	UserId int `json:"user_index"`
	Type   int `json:"type"`
}

type DemoRecordParsedEventHuskRage struct {
	Tick int `json:"tick"`

	UserId int `json:"user_index"`
}

type DemoRecordParsedPlayerEvents struct {
	ConnectionLog []*DemoRecordParsedEventConnection `json:"connection_log"`

	Perks  []*DemoRecordParsedEventPerkChange `json:"perks"`
	Kills  []*DemoRecordParsedEventKill       `json:"kills"`
	Buffs  []*DemoRecordParsedEventBuff       `json:"buffs"`
	Deaths []*DemoRecordParsedEventDeath      `json:"deaths"`

	HuskRages []*DemoRecordParsedEventHuskRage `json:"husk_rages"`

	HealthChanges []*DemoRecordParsedEventHpChange `json:"hp_changes"`
}

type DemoRecordParsedWave struct {
	Wave    int `json:"wave"`
	Attempt int `json:"attempt"`

	StartTick int `json:"start_tick"`
	EndTick   int `json:"end_tick"`
}

type DemoRecordParsedEventZedsLeft struct {
	Tick int `json:"tick"`

	ZedsLeft int `json:"zeds_left"`
}

type DemoRecordParsedWaveEvents struct {
	Waves []*DemoRecordParsedWave `json:"waves"`

	ZedsLeft []*DemoRecordParsedEventZedsLeft `json:"zeds_left"`
	ZedTimes []*DemoRecordParsedZedtime       `json:"zed_times"`
}

type DemoRecordParsed struct {
	Version   byte `json:"protocol_version"`
	SessionId int  `json:"session_id"`

	StartTick int `json:"start_tick"`
	EndTick   int `json:"end_tick"`

	Players []*DemoRecordParsedPlayer `json:"players"`

	WaveEvents   *DemoRecordParsedWaveEvents   `json:"wave_events"`
	PlayerEvents *DemoRecordParsedPlayerEvents `json:"player_events"`
}

func (raw *DemoRecordRaw) ToParsed() (*DemoRecordParsed, error) {
	parsedDemo := DemoRecordParsed{
		Version:   raw.Header.Version,
		SessionId: raw.Header.SessionId,

		Players: []*DemoRecordParsedPlayer{},
		WaveEvents: &DemoRecordParsedWaveEvents{
			Waves:    []*DemoRecordParsedWave{},
			ZedsLeft: []*DemoRecordParsedEventZedsLeft{},
			ZedTimes: []*DemoRecordParsedZedtime{},
		},
		PlayerEvents: &DemoRecordParsedPlayerEvents{
			ConnectionLog: []*DemoRecordParsedEventConnection{},
			Perks:         []*DemoRecordParsedEventPerkChange{},
			Kills:         []*DemoRecordParsedEventKill{},
			Buffs:         []*DemoRecordParsedEventBuff{},
			Deaths:        []*DemoRecordParsedEventDeath{},
			HuskRages:     []*DemoRecordParsedEventHuskRage{},
			HealthChanges: []*DemoRecordParsedEventHpChange{},
		},
	}

	if len(raw.Events) > 0 {
		parsedDemo.StartTick = raw.Events[0].Tick
		parsedDemo.EndTick = raw.Events[len(raw.Events)-1].Tick
	}

	parsedDemo.Players = raw.parsePlayers()

	waves, err := raw.parseWaves()
	if err != nil {
		return nil, err
	}

	parsedDemo.WaveEvents.Waves = waves
	parsedDemo.WaveEvents.ZedTimes = raw.parseZedtimeEvents()

	for i := range raw.Events {
		event := raw.Events[i]

		if event.Type == byte(PlayerJoin) || event.Type == byte(PlayerDisconnect) {
			parsedDemo.PlayerEvents.ConnectionLog =
				append(parsedDemo.PlayerEvents.ConnectionLog, &DemoRecordParsedEventConnection{
					Tick:   event.Tick,
					UserId: int(event.Data["user_id"].(byte)),
					Type:   int(event.Type),
				})
		} else if event.Type == byte(PlayerPerk) {
			parsedDemo.PlayerEvents.Perks =
				append(parsedDemo.PlayerEvents.Perks, &DemoRecordParsedEventPerkChange{
					Tick:   event.Tick,
					UserId: int(event.Data["user_id"].(byte)),
					Perk:   int(event.Data["perk"].(byte)),
				})
		} else if event.Type == byte(PlayerDied) {
			parsedDemo.PlayerEvents.Deaths =
				append(parsedDemo.PlayerEvents.Deaths, &DemoRecordParsedEventDeath{
					Tick:   event.Tick,
					UserId: int(event.Data["user_id"].(byte)),
					Cause:  int(event.Data["cause"].(byte)),
				})
		} else if event.Type == byte(GlobalZedsLeft) {
			parsedDemo.WaveEvents.ZedsLeft =
				append(parsedDemo.WaveEvents.ZedsLeft, &DemoRecordParsedEventZedsLeft{
					Tick:     event.Tick,
					ZedsLeft: event.Data["zeds_left"].(int),
				})
		} else if event.Type == byte(EventKill) {
			parsedDemo.PlayerEvents.Kills =
				append(parsedDemo.PlayerEvents.Kills, &DemoRecordParsedEventKill{
					Tick:   event.Tick,
					UserId: int(event.Data["user_id"].(byte)),
					Zed:    int(event.Data["zed"].(byte)),
				})
		} else if event.Type == byte(EventBuffs) {
			parsedDemo.PlayerEvents.Buffs =
				append(parsedDemo.PlayerEvents.Buffs, &DemoRecordParsedEventBuff{
					Tick:     event.Tick,
					UserId:   int(event.Data["user_id"].(byte)),
					MaxBuffs: int(event.Data["max_buffs"].(byte)),
				})
		} else if event.Type == byte(EventHpChange) {
			parsedDemo.PlayerEvents.HealthChanges =
				append(parsedDemo.PlayerEvents.HealthChanges, &DemoRecordParsedEventHpChange{
					Tick:   event.Tick,
					UserId: int(event.Data["user_id"].(byte)),
					Health: event.Data["health"].(int),
					Armor:  int(event.Data["armor"].(byte)),
				})
		} else if event.Type == byte(EventHuskRage) {
			parsedDemo.PlayerEvents.HuskRages =
				append(parsedDemo.PlayerEvents.HuskRages, &DemoRecordParsedEventHuskRage{
					Tick:   event.Tick,
					UserId: int(event.Data["user_id"].(byte)),
				})
		}
	}

	return &parsedDemo, nil
}

func (raw *DemoRecordRaw) parsePlayers() []*DemoRecordParsedPlayer {
	players := []*DemoRecordParsedPlayer{}

	events := filterEventsByType(raw.Events, byte(PlayerJoin))

	unique := map[int]*DemoRecordParsedPlayer{}

	for i := range events {
		event := events[i]

		player := DemoRecordParsedPlayer{
			UserId:   int(event.Data["user_id"].(byte)),
			UserType: int(event.Data["user_type"].(byte)),
			UniqueId: event.Data["unique_id"].(string),
		}

		unique[player.UserId] = &player
	}

	for _, item := range unique {
		players = append(players, item)
	}

	return players
}

func (raw *DemoRecordRaw) parseWaves() ([]*DemoRecordParsedWave, error) {
	waves := []*DemoRecordParsedWave{}

	waveStart := -1
	waveEnd := -1

	attempts := map[int]int{}

	appendWave := func(startIdx, endIdx int) {
		if startIdx < 0 || endIdx < 0 {
			fmt.Printf("invalid range [%v;%v]\n", startIdx, endIdx)
			return
		}

		wave := int(raw.Events[startIdx].Data["wave"].(byte))

		waves = append(waves, &DemoRecordParsedWave{
			Wave:      wave,
			Attempt:   attempts[wave],
			StartTick: raw.Events[startIdx].Tick,
			EndTick:   raw.Events[endIdx].Tick,
		})
	}

	for i := range raw.Events {
		event := raw.Events[i]

		if event.Type == byte(GlobalWaveStart) {
			waveStart = i

			wave := int(event.Data["wave"].(byte))
			attempts[wave] += 1
		} else if event.Type == byte(GlobalWaveEnd) {
			waveEnd = i

			if waveEnd < waveStart {
				return nil, fmt.Errorf("waveEnd < waveStart at pos %v", waveEnd)
			}

			appendWave(waveStart, waveEnd)
		}
	}

	// Detect if last wave is not finished
	if waveEnd < waveStart {
		appendWave(waveStart, len(raw.Events)-1)
	}

	return waves, nil
}

func (raw *DemoRecordRaw) parseZedtimeEvents() []*DemoRecordParsedZedtime {
	items := []*DemoRecordParsedZedtime{}

	zedtimeEvents := filterEventsByType(raw.Events, byte(GlobalZedTime))
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

		item := DemoRecordParsedZedtime{
			Duration:     float64(zedtimeEvents[rangeEndIndex].Tick+300-zedtimeEvents[rangeStartIndex].Tick) / 100,
			ExtendsCount: rangeEndIndex - rangeStartIndex,
		}

		for j := rangeStartIndex; j <= rangeEndIndex; j++ {
			item.Ticks = append(item.Ticks, zedtimeEvents[j].Tick)
		}

		item.Ticks = append(item.Ticks, zedtimeEvents[rangeEndIndex].Tick+300)
		item.StartTick = item.Ticks[0]
		item.EndTick = item.Ticks[len(item.Ticks)-1]

		items = append(items, &item)
	}

	return items
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
