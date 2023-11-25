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

// @Summary Get user perks analytics
// @Tags 	Analytics
// @Produce json
// @Param   body body 		UserPerksAnalyticsRequest true "Body"
// @Success 200 {object} 	UserPerksAnalyticsResponse
// @Router /analytics/users/perks [post]
func (c *controller) getPerksAnalytics(ctx *gin.Context) {
	var req UserPerksAnalyticsRequest
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	res, err := c.service.GetPerksAnalytics(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// @Summary Get user perk playtime histogram
// @Tags 	Analytics
// @Produce json
// @Param   body body 		UserPerkHistRequest true "Body"
// @Success 200 {object} 	PlayTimeHist
// @Router /analytics/users/perks/playtime [post]
func (c *controller) getPlaytimeHist(ctx *gin.Context) {
	var req UserPerkHistRequest
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	res, err := c.service.getPlaytimeHist(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// @Summary Get user perk accuracy histogram
// @Tags 	Analytics
// @Produce json
// @Param   body body 		UserPerkHistRequest true "Body"
// @Success 200 {object} 	AccuracyHist
// @Router /analytics/users/perks/accuracy [post]
func (c *controller) getAccuracyHist(ctx *gin.Context) {
	var req UserPerkHistRequest
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	res, err := c.service.getAccuracyHist(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// @Summary Get user teammates
// @Tags 	Analytics
// @Produce json
// @Param   body body 		GetTeammatesRequest true "Body"
// @Success 200 {object} 	GetTeammatesResponse
// @Router /analytics/users/teammates [post]
func (c *controller) getTeammates(ctx *gin.Context) {
	var req GetTeammatesRequest
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	res, err := c.service.getTeammates(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}
