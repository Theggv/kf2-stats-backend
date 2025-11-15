package filter

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type matchesController struct {
	service *MatchesFilterService
}

// @Summary Get matches by filter
// @Tags 	Match
// @Produce json
// @Param   filter body 	FilterMatchesRequest true "Get matches by filter"
// @Success 200 {array} 	FilterMatchesResponse
// @Router /matches/filter/new [post]
func (c *matchesController) filter(ctx *gin.Context) {
	var req FilterMatchesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	res, err := c.service.Filter(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}
