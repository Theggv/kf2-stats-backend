package models

import "time"

// Composite primary key by 4 columns
type PlayerStats struct {
	PlayerId  int
	SessionId int
	Wave      int
	Attempt   int

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
