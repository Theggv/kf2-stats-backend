package models

import "time"

type Match struct {
	Session MatchSession `json:"session"`

	Details MatchDetails `json:"details"`

	Metadata SessionMetadata `json:"metadata"`
}

type MatchSession struct {
	Id int `json:"id"`

	ServerId int `json:"server_id"`
	MapId    int `json:"map_id"`

	Mode       GameMode       `json:"mode"`
	Length     int            `json:"length"`
	Difficulty GameDifficulty `json:"diff"`

	Status GameStatus `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type MatchDetails struct {
	Map    *MatchMap    `json:"map,omitempty"`
	Server *MatchServer `json:"server,omitempty"`

	GameData      *GameData      `json:"game_data,omitempty"`
	ExtraGameData *ExtraGameData `json:"extra_data,omitempty"`

	LiveData *MatchLiveData `json:"live_data,omitempty"`

	UserData *MatchUserData `json:"user_data,omitempty"`
}

type MatchMap struct {
	Name    string  `json:"name"`
	Preview *string `json:"preview,omitempty"`
}

type MatchServer struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type MatchLiveData struct {
	Players    []*MatchPlayer `json:"players"`
	Spectators []*MatchPlayer `json:"spectators"`
}

type MatchPlayer struct {
	Profile *UserProfile `json:"profile"`

	Perk     Perk `json:"perk"`
	Level    int  `json:"level"`
	Prestige int  `json:"prestige"`

	Health int `json:"health"`
	Armor  int `json:"armor"`
}

type MatchUserData struct {
	LastSeen *time.Time `json:"last_seen"`

	Perks []int `json:"perks"`

	Stats *MatchUserDataStats `json:"stats,omitempty"`
}

type MatchUserDataStats struct {
	DamageDealt int `json:"damage_dealt"`
}
