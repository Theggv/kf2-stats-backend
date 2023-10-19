package stats

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type statsController struct {
	service *StatsService
}

// @Summary Creates wave stats
// @Tags 	Stats
// @Produce json
// @Param   stats body    	CreateWaveStatsRequest true "Stats JSON"
// @Success 201
// @Router /stats/wave [post]
func (c *statsController) createWaveStats(ctx *gin.Context) {
	var req CreateWaveStatsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	err := c.service.CreateWaveStats(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
