package stats

import "time"

type Perk = int

const (
	Berserker Perk = iota + 1
	Commando
	Medic
	Sharpshooter
	Gunslinger
	Support
	Swat
	Demolitionist
	Firebug
	Survivalist
)

type ZedCounter struct {
	Cyst         int `json:"cyst"`
	AlphaClot    int `json:"alpha_clot"`
	Slasher      int `json:"slasher"`
	Stalker      int `json:"stalker"`
	Crawler      int `json:"crawler"`
	Gorefast     int `json:"gorefast"`
	Rioter       int `json:"rioter"`
	EliteCrawler int `json:"elite_crawler"`
	Gorefiend    int `json:"gorefiend"`

	Siren int `json:"siren"`
	Bloat int `json:"bloat"`
	Edar  int `json:"edar"`
	Husk  int `json:"husk"`

	Scrake int `json:"scrake"`
	FP     int `json:"fp"`
	QP     int `json:"qp"`
	Boss   int `json:"boss"`
}

// Composite primary key by 4 columns
type WavePlayerStats struct {
	Id        int `json:"id"`
	SessionId int `json:"session_id"`
	PlayerId  int `json:"player_id"`
	Wave      int `json:"wave"`
	Attempt   int `json:"attempt"`

	Perk Perk `json:"perk"`

	ShotsFired int `json:"shots_fired"`
	ShotsHit   int `json:"shots_hit"`
	ShotsHS    int `json:"shots_hs"`

	DoshEarned int `json:"dosh_earned"`

	HealsGiven    int `json:"heals_given"`
	HealsReceived int `json:"heals_recv"`

	DamageDealt int `json:"damage_dealt"`
	DamageTaken int `json:"damage_taken"`

	CreatedAt time.Time `json:"created_at"`
}
