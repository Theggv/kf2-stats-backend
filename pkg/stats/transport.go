package stats

type CreateStatsRequestKills struct {
	Cyst         int `json:"cyst"`
	AlphaClot    int `json:"alpha_clot"`
	Slasher      int `json:"slasher"`
	Stalker      int `json:"stalker"`
	Crawler      int `json:"crawler"`
	Gorefast     int `json:"gorefast"`
	Rioter       int `json:"rioter"`
	EliteCrawler int `json:"elite_crawler"`
	Gorefiend    int `json:"gorefiend"`

	Siren        int `json:"siren"`
	Bloat        int `json:"bloat"`
	Edar         int `json:"edar"`
	HuskNormal   int `json:"husk_n"`
	HuskBackpack int `json:"husk_b"`
	HuskRages    int `json:"husk_r"`

	Scrake int `json:"scrake"`
	FP     int `json:"fp"`
	QP     int `json:"qp"`
	Boss   int `json:"boss"`
}

type CreateStatsRequest struct {
	SessionId int `json:"session_id"`
	PlayerId  int `json:"player_id"`
	Wave      int `json:"wave"`
	Attempt   int `json:"attempt"`

	Perk Perk `json:"perk"`

	ShotsFired int `json:"shots_fired"`
	ShotsHit   int `json:"shots_hit"`
	ShotsHS    int `json:"shots_hs"`

	Kills CreateStatsRequestKills `json:"kills"`

	DoshEarned int `json:"dosh_earned"`

	HealsGiven    int `json:"heals_given"`
	HealsReceived int `json:"heals_recv"`

	DamageDealt int `json:"damage_dealt"`
	DamageTaken int `json:"damage_taken"`
}
