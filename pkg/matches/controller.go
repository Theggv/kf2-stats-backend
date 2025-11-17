package matches

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type controller struct {
	service *MatchesService
}

// @Summary Get match by id
// @Tags 	Match
// @Produce json
// @Param   id path   	 	int true "Session id"
// @Success 200 {object} 	Match
// @Router /matches/{id} [get]
func (c *controller) getById(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	item, err := c.service.GetById(id)
	if err != nil {
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// @Summary Get match live data
// @Tags 	Match
// @Produce json
// @Param   id path   	 	int true "Session id"
// @Success 200 {object} 	GetMatchLiveDataResponse
// @Router /matches/{id}/live [get]
func (c *controller) getMatchLiveData(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	item, err := c.service.GetMatchLiveData(id)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// @Summary Get last server session
// @Tags 	Match
// @Produce json
// @Param   id path   	 	int true "Server id"
// @Success 200 {object} 	Match
// @Router /matches/server/{id} [get]
func (c *controller) getLastServerMatch(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	item, err := c.service.GetLastServerMatch(id)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// @Summary Get match waves
// @Tags 	Match
// @Produce json
// @Param   id path   	 	int true "Session id"
// @Success 200 {object} 	GetMatchWavesResponse
// @Router /matches/{id}/waves [get]
func (c *controller) getMatchWaves(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	item, err := c.service.GetMatchWaves(id)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// @Summary Get wave players stats
// @Tags 	Match
// @Produce json
// @Param   id path   	 	int true "Session id"
// @Param   userId path   	 	int true "User id"
// @Success 200 {object} 	GetMatchPlayerStatsResponse
// @Router /matches/{id}/user/{userId}/stats [get]
func (c *controller) getMatchPlayerStats(ctx *gin.Context) {
	sessionId, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	userId, err := strconv.Atoi(ctx.Params.ByName("userId"))

	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	item, err := c.service.GetMatchPlayerStats(sessionId, userId)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// @Summary Get wave players stats
// @Tags 	Match
// @Produce json
// @Param   id path   	 	int true "Session id"
// @Success 200 {object} 	GetMatchAggregatedStatsResponse
// @Router /matches/{id}/summary [get]
func (c *controller) getMatchAggregatedStats(ctx *gin.Context) {
	sessionId, err := strconv.Atoi(ctx.Params.ByName("id"))

	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	item, err := c.service.GetMatchAggregatedStats(sessionId)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, item)
}
