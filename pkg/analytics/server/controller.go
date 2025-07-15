package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type controller struct {
	service *ServerAnalyticsService
}

// @Summary Get played session count for certain period grouped by time period
// @Tags 	Analytics
// @Produce json
// @Param   body body 		SessionCountRequest true "Body"
// @Success 200 {object} 	SessionCountResponse
// @Router /analytics/server/session/count [post]
func (c *controller) getSessionCount(ctx *gin.Context) {
	var req SessionCountRequest
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	items, err := c.service.GetSessionCount(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, &SessionCountResponse{
		Items: *items,
	})
}

// @Summary Get server usage in minutes for certain period grouped by time period
// @Tags 	Analytics
// @Produce json
// @Param   body body 		UsageInMinutesRequest true "Body"
// @Success 200 {object} 	UsageInMinutesResponse
// @Router /analytics/server/usage [post]
func (c *controller) getUsageInMinutes(ctx *gin.Context) {
	var req UsageInMinutesRequest
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	items, err := c.service.GetUsageInMinutes(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, &UsageInMinutesResponse{
		Items: *items,
	})
}

// @Summary Get server online for certain period grouped by time period
// @Tags 	Analytics
// @Produce json
// @Param   body body 		PlayersOnlineRequest true "Body"
// @Success 200 {object} 	PlayersOnlineResponse
// @Router /analytics/server/online [post]
func (c *controller) getPlayersOnline(ctx *gin.Context) {
	var req PlayersOnlineRequest
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	items, err := c.service.GetPlayersOnline(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, &PlayersOnlineResponse{
		Items: *items,
	})
}

// @Summary Get popular servers by sessions count
// @Tags 	Analytics
// @Produce json
// @Success 200 {object} 	PopularServersResponse
// @Router /analytics/server/popular [get]
func (c *controller) getPopularServers(ctx *gin.Context) {
	res, err := c.service.GetPopularServers()
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// @Summary Get number of currently online players
// @Tags 	Analytics
// @Produce json
// @Success 200 {object} 	TotalOnlineResponse
// @Router /analytics/server/current-online [get]
func (c *controller) getCurrentOnline(ctx *gin.Context) {
	res, err := c.service.GetCurrentOnline()
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}
