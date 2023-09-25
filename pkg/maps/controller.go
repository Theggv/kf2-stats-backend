package maps

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type mapsController struct {
	service *MapsService
}

func (c *mapsController) add(ctx *gin.Context) {
	var req AddMapRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	id, err := c.service.Create(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, AddMapResponse{
		Id: id,
	})
}

func (c *mapsController) getById(ctx *gin.Context) {
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

func (c *mapsController) getByPattern(ctx *gin.Context) {
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

func (c *mapsController) updatePreview(ctx *gin.Context) {
	var req UpdatePreviewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	err := c.service.UpdatePreview(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
