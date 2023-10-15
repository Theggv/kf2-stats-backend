package stats

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type statsController struct {
	service *StatsService
}

// @Summary Creates stats
// @Tags 	Stats
// @Produce json
// @Param   stats body    	CreateStatsRequest true "Stats JSON"
// @Success 201
// @Router /stats [post]
func (c *statsController) create(ctx *gin.Context) {
	var req CreateStatsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	err := c.service.Create(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
