package maps

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type mapsController struct {
	service *MapsService
}

// @Summary Creates a new map
// @Tags 	Maps
// @Produce json
// @Param   map body    AddMapRequest true "Map JSON"
// @Success 201 {object} AddMapResponse
// @Router /maps [post]
func (c *mapsController) add(ctx *gin.Context) {
	var req AddMapRequest
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

	ctx.JSON(http.StatusCreated, AddMapResponse{
		Id: id,
	})
}

// @Summary Get map by id
// @Tags 	Maps
// @Produce json
// @Param   id path   	 int true "Map id"
// @Success 200 {object} Map
// @Router /maps/{id} [get]
func (c *mapsController) getById(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	item, err := c.service.GetById(id)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// @Summary Get maps by pattern
// @Tags 	Maps
// @Produce json
// @Param   pattern query string false "Get maps by pattern"
// @Success 200 {array} Map
// @Router /maps [get]
func (c *mapsController) getByPattern(ctx *gin.Context) {
	pattern := ctx.Query("pattern")

	items, err := c.service.GetByPattern(pattern)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, GetByPatternResponse{
		Items: items,
	})
}

// @Summary Update map preview
// @Tags 	Maps
// @Produce json
// @Param   body body UpdatePreviewRequest true "Body"
// @Success 200
// @Router /maps/preview [put]
func (c *mapsController) updatePreview(ctx *gin.Context) {
	var req UpdatePreviewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	err := c.service.UpdatePreview(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
