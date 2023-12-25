package leaderboards

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type controller struct {
	service *LeaderBoardsService
}

// @Summary Get leaderboard
// @Tags 	Leaderboards
// @Produce json
// @Param   body body 		LeaderBoardsRequest true "Body"
// @Success 200 {object} 	LeaderBoardsResponse
// @Router /leaderboards/ [post]
func (c *controller) getLeaderBoard(ctx *gin.Context) {
	var req LeaderBoardsRequest
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	res, err := c.service.getLeaderBoard(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}
