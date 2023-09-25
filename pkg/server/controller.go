package server

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ServerController struct {
	service *ServerService
}

func newServerController(db *sql.DB) *ServerController {
	return &ServerController{
		service: NewServerService(db),
	}
}

func (c *ServerController) add(ctx *gin.Context) {
	var req AddServerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	id, err := c.service.CreateServer(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, AddServerResponse{
		Id: id,
	})
}

func (c *ServerController) getById(ctx *gin.Context) {
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

func (c *ServerController) getByPattern(ctx *gin.Context) {
	pattern := ctx.Query("pattern")

	items, err := c.service.GetByPattern(pattern)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, GetByPatternResponse{
		Items: items,
	})
}

func (c *ServerController) updateName(ctx *gin.Context) {
	var req UpdateNameRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	err := c.service.UpdateName(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
