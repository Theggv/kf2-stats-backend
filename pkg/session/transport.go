package session

import (
	"time"

	"github.com/theggv/kf2-stats-backend/pkg/common/models"
	"github.com/theggv/kf2-stats-backend/pkg/maps"
	"github.com/theggv/kf2-stats-backend/pkg/server"
)

type CreateSessionRequest struct {
	ServerId int `json:"server_id" binding:"required"`
	MapId    int `json:"map_id" binding:"required"`

	Mode       models.GameMode       `json:"mode" binding:"required"`
	Length     int                   `json:"length" binding:"required"`
	Difficulty models.GameDifficulty `json:"diff" binding:"required"`
}

type CreateSessionResponse struct {
	Id int `json:"id"`
}

type FilterSessionsRequest struct {
	ServerId []int `json:"server_id"`
	MapId    []int `json:"map_id"`

	Mode       models.GameMode       `json:"mode"`
	Length     models.GameLength     `json:"length"`
	Difficulty models.GameDifficulty `json:"diff"`

	IncludeServer bool `json:"include_server"`
	IncludeMap    bool `json:"include_map"`

	Pager models.PaginationRequest `json:"pager"`
}

type FilterSessionsResponse struct {
	Items    []Session                 `json:"items"`
	Metadata models.PaginationResponse `json:"metadata"`
}

type UpdateStatusRequest struct {
	Id     int `json:"id" binding:"required"`
	Status int `json:"status" binding:"required"`
}

type GameData struct {
	MaxPlayers    int `json:"max_players"`
	PlayersOnline int `json:"players_online"`
	PlayersAlive  int `json:"players_alive"`

	Wave         int  `json:"wave"`
	IsTraderTime bool `json:"is_trader_time"`
	ZedsLeft     int  `json:"zeds_left"`
}

type LiveMatch struct {
	SessionId int `json:"session_id"`

	Mode       models.GameMode       `json:"mode"`
	Length     models.GameLength     `json:"length"`
	Difficulty models.GameDifficulty `json:"diff"`

	Status *models.GameStatus `json:"status"`

	Map    *maps.Map      `json:"map"`
	Server *server.Server `json:"server"`

	GameData GameData           `json:"game_data"`
	CDData   *models.CDGameData `json:"cd_data"`

	StartedAt *time.Time `json:"started_at"`
}

type GetLiveMatchesResponse struct {
	Items []LiveMatch `json:"items"`
}

type UpdateGameDataRequest struct {
	SessionId int `json:"session_id"`

	GameData GameData           `json:"game_data"`
	CDData   *models.CDGameData `json:"cd_data"`
}
