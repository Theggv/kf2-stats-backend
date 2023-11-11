package maps

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type controller struct {
	service *MapAnalyticsService
}

// @Summary Get maps analytics
// @Tags 	Analytics
// @Produce json
// @Param   body body 		MapAnalyticsRequest true "Body"
// @Success 200 {object} 	MapAnalyticsResponse
// @Router /analytics/maps [post]
func (c *controller) getMapAnalytics(ctx *gin.Context) {
	var req MapAnalyticsRequest
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	items, err := c.service.GetMapAnalytics(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, &MapAnalyticsResponse{
		Items: *items,
	})
}
