package matches

import (
	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/stats"
)

type FilterMatchesRequest struct {
	ServerId []int               `json:"server_id"`
	MapId    []int               `json:"map_id"`
	Status   []models.GameStatus `json:"status"`

	Mode       *models.GameMode       `json:"mode"`
	Length     *models.GameLength     `json:"length"`
	Difficulty *models.GameDifficulty `json:"diff"`

	IncludeServer   *bool `json:"include_server"`
	IncludeMap      *bool `json:"include_map"`
	IncludeGameData *bool `json:"include_game_data"`
	IncludeCDData   *bool `json:"include_cd_data"`
	IncludePlayers  *bool `json:"include_players"`

	ReverseOrder *bool                    `json:"reverse_order"`
	Pager        models.PaginationRequest `json:"pager"`
}

type FilterMatchesResponse struct {
	Items    []*Match                  `json:"items"`
	Metadata models.PaginationResponse `json:"metadata"`
}

type GetLiveMatchesResponse struct {
	Items []Match `json:"items"`
}

type GetMatchWavesResponse struct {
	Waves []*MatchWave          `json:"waves"`
	Users []*models.UserProfile `json:"users"`
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
}

type GetMatchWaveStatsResponse struct {
	Players []PlayerWaveStats `json:"players"`
}

type GetMatchPlayerStatsResponse struct {
	Waves []PlayerWaveStats `json:"waves"`
}

type AggregatedPlayerStats struct {
	UserId int `json:"user_id"`

	// in seconds
	PlayTime int `json:"play_time"`

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

	Kills      int `json:"kills"`
	LargeKills int `json:"large_kills"`
	HuskRages  int `json:"husk_r"`
}

type GetMatchAggregatedStatsResponse struct {
	Players []AggregatedPlayerStats `json:"players"`
}

type GetMatchLiveDataResponsePlayer struct {
	Id   int    `json:"id"`
	Name string `json:"name"`

	ProfileUrl *string `json:"profile_url"`
	Avatar     *string `json:"avatar"`

	Perk     models.Perk `json:"perk"`
	Level    int         `json:"level"`
	Prestige int         `json:"prestige"`

	Health int `json:"health"`
	Armor  int `json:"armor"`

	AuthId      string          `json:"-"`
	AuthType    models.AuthType `json:"-"`
	IsSpectator bool            `json:"-"`
}

type GetMatchLiveDataResponse struct {
	Status models.GameStatus `json:"status"`

	GameData models.GameData    `json:"game_data"`
	CDData   *models.CDGameData `json:"cd_data"`

	Players    []*GetMatchLiveDataResponsePlayer `json:"players"`
	Spectators []*GetMatchLiveDataResponsePlayer `json:"spectators"`
}
