package session

import (
	"net/http"
	"strconv"

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
		return
	}

	id, err := c.service.CreateSession(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, CreateSessionResponse{
		Id: id,
	})
}

// @Summary Get session by id
// @Tags 	Session
// @Produce json
// @Param   id path   	 	int true "Session id"
// @Success 200 {object} 	Session
// @Router /sessions/{id} [get]
func (c *sessionController) getById(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	item, err := c.service.GetById(id)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// @Summary Get sessions by filter
// @Tags 	Session
// @Produce json
// @Param   filter body 	FilterSessionsRequest true "Get sessions by filter"
// @Success 200 {array} 	FilterSessionsResponse
// @Router /sessions/filter [post]
func (c *sessionController) filter(ctx *gin.Context) {
	var req FilterSessionsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	res, err := c.service.FilterSessions(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// @Summary Update session status
// @Tags 	Session
// @Produce json
// @Param   body body 		UpdateStatusRequest true "Body"
// @Success 200
// @Router /sessions/name [put]
func (c *sessionController) updateStatus(ctx *gin.Context) {
	var req UpdateStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	err := c.service.UpdateStatus(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
