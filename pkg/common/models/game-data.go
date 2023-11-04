package models

type GameData struct {
	MaxPlayers    int `json:"max_players"`
	PlayersOnline int `json:"players_online"`
	PlayersAlive  int `json:"players_alive"`

	Wave         int  `json:"wave"`
	IsTraderTime bool `json:"is_trader_time"`
	ZedsLeft     int  `json:"zeds_left"`
}

type PlayerLiveData struct {
	AuthId   string   `json:"auth_id"`
	AuthType AuthType `json:"auth_type"`
	Name     string   `json:"name"`

	Perk     Perk `json:"perk"`
	Level    int  `json:"level"`
	Prestige int  `json:"prestige"`

	Health int `json:"health"`
	Armor  int `json:"armor"`

	IsSpectator bool `json:"is_spectator"`
}
