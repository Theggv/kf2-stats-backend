package models

import "time"

type Perk = int

const (
	Berserker Perk = iota
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

// Composite primary key by 4 columns
type PlayerStats struct {
	SessionId int
	PlayerId  int
	Wave      int
	Attempt   int

	Perk Perk

	Accuracy   float32
	HSAccuracy float32

	TrashKills  int
	MediumKills int
	ScrakeKills int
	FPKills     int
	MiniFPKills int
	BossKills   int

	HuskNormalKills   int
	HuskBackpackKills int
	HuskRages         int

	DamageDealt int
	DamageTaken int

	CreatedAt time.Time
}
