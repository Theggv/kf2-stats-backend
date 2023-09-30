package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type serverController struct {
	service *ServerService
}

// @Summary Creates a new server
// @Tags 	Server
// @Produce json
// @Param   server body    	AddServerRequest true "Server JSON"
// @Success 201 {object} 	AddServerResponse
// @Router /servers [post]
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

// @Summary Get server by id
// @Tags 	Server
// @Produce json
// @Param   id path   	 	int true "Server id"
// @Success 200 {object} 	Server
// @Router /servers/{id} [get]
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

// @Summary Get servers by pattern
// @Tags 	Server
// @Produce json
// @Param   pattern query 	string false "Get servers by pattern"
// @Success 200 {array} 	Server
// @Router /servers [get]
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

// @Summary Update server name
// @Tags 	Server
// @Produce json
// @Param   body body 		UpdateNameRequest true "Body"
// @Success 200
// @Router /servers/name [put]
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
