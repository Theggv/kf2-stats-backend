package stats

import (
	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

type CreateWaveStatsRequestPlayer struct {
	UserName     string          `json:"user_name"`
	UserAuthId   string          `json:"user_auth_id"`
	UserAuthType models.AuthType `json:"user_auth_type"`

	Perk     models.Perk `json:"perk"`
	Level    int         `json:"level"`
	Prestige int         `json:"prestige"`

	IsDead bool `json:"is_dead"`

	ShotsFired int `json:"shots_fired"`
	ShotsHit   int `json:"shots_hit"`
	ShotsHS    int `json:"shots_hs"`

	Kills ZedCounter `json:"kills"`

	HuskBackpackKills int `json:"husk_b"`
	HuskRages         int `json:"husk_r"`

	Injuredby ZedCounter `json:"injured_by"`

	DoshEarned int `json:"dosh_earned"`

	HealsGiven    int `json:"heals_given"`
	HealsReceived int `json:"heals_recv"`

	DamageDealt int `json:"damage_dealt"`
	DamageTaken int `json:"damage_taken"`

	ZedTimeCount  int     `json:"zedtime_count"`
	ZedTimeLength float32 `json:"zedtime_length"`

	RequestHealing int `json:"request_healing"`
	RequestDosh    int `json:"request_dosh"`
	RequestHelp    int `json:"request_help"`
	TauntZeds      int `json:"taunt_zeds"`
	FollowMe       int `json:"follow_me"`
	GetToTheTrader int `json:"get_to_the_trader"`
	Affirmative    int `json:"affirmative"`
	Negative       int `json:"negative"`
	ThankYou       int `json:"thank_you"`
}

type CreateWaveStatsRequest struct {
	SessionId int `json:"session_id" binding:"required"`
	Wave      int `json:"wave"`
	Length    int `json:"wave_length"`

	CDData *models.CDGameData `json:"cd_data"`

	Players []CreateWaveStatsRequestPlayer `json:"players"`
}
