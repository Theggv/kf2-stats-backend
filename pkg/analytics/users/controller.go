package users

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type controller struct {
	service *UserAnalyticsService
}

// @Summary Get user analytics
// @Tags 	Analytics
// @Produce json
// @Param   body body 		UserAnalyticsRequest true "Body"
// @Success 200 {object} 	UserAnalyticsResponse
// @Router /analytics/users [post]
func (c *controller) getUserAnalytics(ctx *gin.Context) {
	var req UserAnalyticsRequest
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	res, err := c.service.GetUserAnalytics(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}
