package stats

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type statsController struct {
	service *StatsService
}

// @Summary Creates player wave stats
// @Tags 	Stats
// @Produce json
// @Param   stats body    	CreateWavePlayerStatsRequest true "Stats JSON"
// @Success 201
// @Router /stats/wave/player [post]
func (c *statsController) createWavePlayerStats(ctx *gin.Context) {
	var req CreateWavePlayerStatsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	err := c.service.CreateWavePlayerStats(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

// @Summary Creates cd wave stats
// @Tags 	Stats
// @Produce json
// @Param   stats body    	CreateWaveStatsCDRequest true "Stats JSON"
// @Success 201
// @Router /stats/wave/cd [post]
func (c *statsController) createWaveStatsCD(ctx *gin.Context) {
	var req CreateWaveStatsCDRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	err := c.service.CreateWaveStatsCD(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
