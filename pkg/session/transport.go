package session

import "github.com/theggv/kf2-stats-backend/pkg/common/models"

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

type UpdateGameDataRequest struct {
	SessionId int `json:"session_id"`

	Wave         int  `json:"wave"`
	IsTraderTime bool `json:"is_trader_time"`
	ZedsLeft     int  `json:"zeds_left"`
	PlayersAlive int  `json:"players_alive"`
}
