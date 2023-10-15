package session

import "github.com/theggv/kf2-stats-backend/pkg/common/models"

type CreateSessionRequest struct {
	ServerId int `json:"server_id" binding:"required"`
	MapId    int `json:"map_id" binding:"required"`

	Mode       Mode       `json:"mode" binding:"required"`
	Length     Length     `json:"length" binding:"required"`
	Difficulty Difficulty `json:"diff" binding:"required"`
}

type CreateSessionResponse struct {
	Id int `json:"id"`
}

type FilterSessionsRequest struct {
	ServerId []int `json:"server_id"`
	MapId    []int `json:"map_id"`

	Mode       Mode       `json:"mode"`
	Length     Length     `json:"length"`
	Difficulty Difficulty `json:"diff"`

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
