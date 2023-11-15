package models

type GameData struct {
	MaxPlayers    int `json:"max_players"`
	PlayersOnline int `json:"players_online"`
	PlayersAlive  int `json:"players_alive"`

	Wave         int  `json:"wave"`
	IsTraderTime bool `json:"is_trader_time"`
	ZedsLeft     int  `json:"zeds_left"`
}
