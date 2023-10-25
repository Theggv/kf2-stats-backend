package matches

import (
	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/stats"
)

type FilterMatchesRequest struct {
	ServerId []int `json:"server_id"`
	MapId    []int `json:"map_id"`

	Mode       *models.GameMode       `json:"mode"`
	Length     *models.GameLength     `json:"length"`
	Difficulty *models.GameDifficulty `json:"diff"`
	Status     *models.GameStatus     `json:"status"`

	IncludeServer   bool `json:"include_server"`
	IncludeMap      bool `json:"include_map"`
	IncludeGameData bool `json:"include_game_data"`
	IncludeCDData   bool `json:"include_cd_data"`

	ReverseOrder *bool                    `json:"reverse_order"`
	Pager        models.PaginationRequest `json:"pager"`
}

type FilterMatchesResponse struct {
	Items    []Match                   `json:"items"`
	Metadata models.PaginationResponse `json:"metadata"`
}

type GetMatchWavesResponse struct {
	Waves []MatchWave `json:"waves"`
}

type PlayerWaveStats struct {
	PlayerStatsId int `json:"player_stats_id"`

	ShotsFired int `json:"shots_fired"`
	ShotsHit   int `json:"shots_hit"`
	ShotsHS    int `json:"shots_hs"`

	DoshEarned int `json:"dosh_earned"`

	HealsGiven    int `json:"heals_given"`
	HealsReceived int `json:"heals_recv"`

	DamageDealt int `json:"damage_dealt"`
	DamageTaken int `json:"damage_taken"`

	ZedTimeCount  int     `json:"zedtime_count"`
	ZedTimeLength float32 `json:"zedtime_length"`

	Kills stats.ZedCounter `json:"kills"`

	HuskBackpackKills int `json:"husk_b"`
	HuskRages         int `json:"husk_r"`

	Injuredby stats.ZedCounter `json:"injured_by"`
}

type GetMatchWaveStatsResponse struct {
	Players []PlayerWaveStats `json:"players"`
}

type GetMatchPlayerStatsResponse struct {
	Waves []PlayerWaveStats `json:"waves"`
}
