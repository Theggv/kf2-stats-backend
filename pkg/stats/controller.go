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
// @Param   stats body    	CreateWaveStatsRequest true "Stats JSON"
// @Success 201
// @Router /stats [post]
func (c *statsController) createWaveStats(ctx *gin.Context) {
	var req CreateWaveStatsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	err := c.service.CreateWaveStats(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
