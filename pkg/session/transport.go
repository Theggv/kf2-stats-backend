package session

import (
	"github.com/theggv/kf2-stats-backend/pkg/common/models"
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

type UpdateStatusRequest struct {
	Id     int `json:"id" binding:"required"`
	Status int `json:"status" binding:"required"`
}

type UpdateGameDataRequest struct {
	SessionId int `json:"session_id"`

	GameData GameData           `json:"game_data"`
	CDData   *models.CDGameData `json:"cd_data"`
}
