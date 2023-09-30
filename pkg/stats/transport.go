package stats

type CreateStatsRequest struct {
	SessionId int `json:"session_id"`
	PlayerId  int `json:"player_id"`
	Wave      int `json:"wave"`
	Attempt   int `json:"attempt"`

	Perk Perk `json:"perk"`

	Accuracy   float32 `json:"accuracy"`
	HSAccuracy float32 `json:"hs_accuracy"`

	TrashKills  int `json:"trash_kills"`
	MediumKills int `json:"medium_kills"`
	ScrakeKills int `json:"scrake_kills"`
	FPKills     int `json:"fp_kills"`
	MiniFPKills int `json:"minifp_kills"`
	BossKills   int `json:"boss_kills"`

	HuskNormalKills   int `json:"husk_n"`
	HuskBackpackKills int `json:"husk_b"`
	HuskRages         int `json:"husk_r"`

	DamageDealt int `json:"damage_dealt"`
	DamageTaken int `json:"damage_taken"`
}
