package stats

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type statsController struct {
	service *StatsService
}

func (c *statsController) create(ctx *gin.Context) {
	var req CreateStatsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	err := c.service.CreateStats(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
