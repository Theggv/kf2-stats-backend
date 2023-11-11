package perks

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type controller struct {
	service *PerksAnalyticsService
}

// @Summary Get play time for each perk for certain period
// @Tags 	Analytics
// @Produce json
// @Param   body body 		PerksPlayTimeRequest true "Body"
// @Success 200 {object} 	PerksPlayTimeResponse
// @Router /analytics/perks/playtime [post]
func (c *controller) getPerksPlayTime(ctx *gin.Context) {
	var req PerksPlayTimeRequest
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	items, err := c.service.GetPerksPlayTime(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, &PerksPlayTimeResponse{
		Items: *items,
	})
}

// @Summary Get kills count for each perk for certain period
// @Tags 	Analytics
// @Produce json
// @Param   body body 		PerksKillsRequest true "Body"
// @Success 200 {object} 	PerksKillsResponse
// @Router /analytics/perks/kills [post]
func (c *controller) getPerksKills(ctx *gin.Context) {
	var req PerksKillsRequest
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	items, err := c.service.GetPerksKills(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, &PerksKillsResponse{
		Items: *items,
	})
}
