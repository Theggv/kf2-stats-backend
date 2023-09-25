package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type serverController struct {
	service *ServerService
}

func (c *serverController) add(ctx *gin.Context) {
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

func (c *serverController) getById(ctx *gin.Context) {
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

func (c *serverController) getByPattern(ctx *gin.Context) {
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

func (c *serverController) updateName(ctx *gin.Context) {
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
