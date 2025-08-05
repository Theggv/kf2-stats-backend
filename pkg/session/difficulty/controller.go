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
