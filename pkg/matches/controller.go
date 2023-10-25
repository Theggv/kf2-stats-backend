package matches

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type matchesController struct {
	service *MatchesService
}

// @Summary Get match by id
// @Tags 	Match
// @Produce json
// @Param   id path   	 	int true "Session id"
// @Success 200 {object} 	Match
// @Router /matches/{id} [get]
func (c *matchesController) getById(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	item, err := c.service.getById(id)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
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
func (c *matchesController) getLastServerMatch(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	item, err := c.service.getLastServerMatch(id)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// @Summary Get matches by filter
// @Tags 	Match
// @Produce json
// @Param   filter body 	FilterMatchesRequest true "Get matches by filter"
// @Success 200 {array} 	FilterMatchesResponse
// @Router /matches/filter [post]
func (c *matchesController) filter(ctx *gin.Context) {
	var req FilterMatchesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	res, err := c.service.filter(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// @Summary Get match waves
// @Tags 	Match
// @Produce json
// @Param   id path   	 	int true "Session id"
// @Success 200 {object} 	GetMatchWavesResponse
// @Router /matches/{id}/waves [get]
func (c *matchesController) getMatchWaves(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	item, err := c.service.getMatchWaves(id)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, item)
}

// @Summary Get wave players stats
// @Tags 	Match
// @Produce json
// @Param   id path   	 	int true "Wave id"
// @Success 200 {object} 	GetMatchWaveStatsResponse
// @Router /matches/wave/{id}/stats [get]
func (c *matchesController) getWavePlayersStats(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	item, err := c.service.getWavePlayersStats(id)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
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
func (c *matchesController) getMatchPlayerStats(ctx *gin.Context) {
	sessionId, err := strconv.Atoi(ctx.Params.ByName("id"))
	userId, err := strconv.Atoi(ctx.Params.ByName("userId"))

	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	item, err := c.service.getMatchPlayerStats(sessionId, userId)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, item)
}
