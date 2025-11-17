package filter

import (
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/common/models/filter"
	"github.com/theggv/kf2-stats-backend/pkg/matches"
)

type FilterMatchesRequestIncludes struct {
	ServerData    *bool `json:"server_data"`
	MapData       *bool `json:"map_data"`
	GameData      *bool `json:"game_data"`
	ExtraGameData *bool `json:"extra_game_data"`
	LiveData      *bool `json:"live_data"`
}

type FilterMatchesRequestExtra struct {
	Wave        *filter.AdvancedFilter `json:"wave"`
	Difficulty  *filter.AdvancedFilter `json:"difficulty"`
	MaxMonsters *filter.AdvancedFilter `json:"max_monsters"`
	SpawnCycle  *string                `json:"spawn_cycle"`
	ZedsType    *string                `json:"zeds_type"`
}

type FilterMatchesRequestExclude struct {
	ServerIds []int `json:"server_ids"`
	MapIds    []int `json:"map_ids"`

	Statuses []models.GameStatus `json:"statuses"`
}

type FilterMatchesRequest struct {
	UserIds   []int `json:"user_ids"`
	ServerIds []int `json:"server_ids"`
	MapIds    []int `json:"map_ids"`

	Exclude *FilterMatchesRequestExclude `json:"exclude"`

	Perks    []models.Perk       `json:"perks"`
	Statuses []models.GameStatus `json:"statuses"`

	Mode       *models.GameMode       `json:"mode"`
	Length     *models.GameLength     `json:"length"`
	Difficulty *models.GameDifficulty `json:"diff"`

	From *time.Time `json:"date_from"`
	To   *time.Time `json:"date_to"`

	Includes *FilterMatchesRequestIncludes `json:"includes"`
	Extra    *FilterMatchesRequestExtra    `json:"extra"`

	SortBy models.SortByRequest     `json:"sort_by"`
	Pager  models.PaginationRequest `json:"pager"`

	AuthUser *models.TokenPayload `json:"-"`
}

type FilterMatchesResponse struct {
	Items []*matches.Match `json:"items"`

	Metadata *models.PaginationResponse `json:"metadata"`
}
