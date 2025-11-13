package difficulty

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type controller struct {
	service *DifficultyCalculatorService
}

// @Summary Recalculate all session difficulties
// @Tags 	Difficulty
// @Produce json
// @Success 201
// @Router /sessions/difficulty/server [post]
func (c *controller) recalculateAll(ctx *gin.Context) {
	err := c.service.RecalculateAll()
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{})
}

// @Summary Recalculate session difficulties by server_id
// @Tags 	Difficulty
// @Produce json
// @Param   id path   	 	int true "Server id"
// @Success 201
// @Router /sessions/difficulty/server/{id} [post]
func (c *controller) recalculateByServerId(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	err = c.service.RecalculateByServerId(id)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{})
}

// @Summary Get session difficulty by session_id
// @Tags 	Difficulty
// @Produce json
// @Param   id path   	 	int true "Session id"
// @Success 200 {object} 	GetSessionDifficultyResponse
// @Router /sessions/difficulty/{id} [get]
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

// @Summary Request to recalculate session difficulty by session_id
// @Tags 	Difficulty
// @Produce json
// @Param   id path   	 	int true "Session id"
// @Success 201
// @Router /sessions/difficulty/{id} [post]
func (c *controller) addToQueue(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	c.service.AddToQueue(id)
	ctx.JSON(http.StatusCreated, gin.H{})
}

// @Summary Check if session is queued for difficulty calculation
// @Tags 	Difficulty
// @Produce json
// @Param   id path   	 	int true "Session id"
// @Success 200
// @Router /sessions/difficulty/{id}/check [get]
func (c *controller) checkIfQueued(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	isInsideQueue := c.service.CheckIfQueued(id)
	ctx.JSON(http.StatusOK, isInsideQueue)
}
