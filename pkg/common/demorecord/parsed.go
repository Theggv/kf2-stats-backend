package demorecord

type DemoRecordParsedPlayer struct {
	UserId   int    `json:"user_id"`
	UserType int    `json:"user_type"`
	UniqueId string `json:"unique_id"`
}

type DemoRecordParsedEventPerkChange struct {
	UserId int `json:"user_id"`
	Perk   int `json:"perk"`
}

type DemoRecordParsedEventDeath struct {
	Tick int `json:"tick"`

	UserId int `json:"user_id"`
	Cause  int `json:"cause"`
}

type DemoRecordParsedZedtime struct {
	StartTick int   `json:"start_tick"`
	EndTick   int   `json:"end_tick"`
	Ticks     []int `json:"ticks"`

	Duration     float32 `json:"duration"`
	ExtendsCount int     `json:"extends_count"`
}

type DemoRecordParsedEventKill struct {
	Tick int `json:"tick"`

	UserId int `json:"user_id"`
	Zed    int `json:"zed"`
}

type DemoRecordParsedEventBuff struct {
	Tick int `json:"tick"`

	UserId   int `json:"user_id"`
	MaxBuffs int `json:"max_buffs"`
}

type DemoRecordParsedEventHpChange struct {
	Tick int `json:"tick"`

	UserId int `json:"user_id"`
	Health int `json:"health"`
	Armor  int `json:"armor"`
}

type DemoRecordParsedEventConnection struct {
	Tick int `json:"tick"`

	UserId int `json:"user_id"`
	Type   int `json:"type"`
}

type DemoRecordParsedEventHuskRage struct {
	Tick int `json:"tick"`

	UserId int `json:"user_id"`
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
