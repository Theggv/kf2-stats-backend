package session

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type sessionController struct {
	service *SessionService
}

// @Summary Creates a new session
// @Tags 	Session
// @Produce json
// @Param   session body    CreateSessionRequest true "Session JSON"
// @Success 201 {object} 	CreateSessionResponse
// @Router /sessions [post]
func (c *sessionController) create(ctx *gin.Context) {
	var req CreateSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	id, err := c.service.Create(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, CreateSessionResponse{
		Id: id,
	})
}

// @Summary Update session status
// @Tags 	Session
// @Produce json
// @Param   body body 		UpdateStatusRequest true "Body"
// @Success 200
// @Router /sessions/status [put]
func (c *sessionController) updateStatus(ctx *gin.Context) {
	var req UpdateStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	err := c.service.UpdateStatus(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

// @Summary Update session game data
// @Tags 	Session
// @Produce json
// @Param   body body 		UpdateGameDataRequest true "Body"
// @Success 200
// @Router /sessions/game-data [put]
func (c *sessionController) updateGameData(ctx *gin.Context) {
	var req UpdateGameDataRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	err := c.service.UpdateGameData(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
