package session

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type sessionController struct {
	service *SessionService
}

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
