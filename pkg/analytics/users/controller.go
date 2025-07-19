package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
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
		return
	}

	res, err := c.service.GetUserAnalytics(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
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
		return
	}

	res, err := c.service.GetPerksAnalytics(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
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
		return
	}

	req.AuthUser, _ = util.GetUserFromCtx(ctx)

	res, err := c.service.getPlaytimeHist(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
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
		return
	}

	req.AuthUser, _ = util.GetUserFromCtx(ctx)

	res, err := c.service.getAccuracyHist(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
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
		return
	}

	req.AuthUser, _ = util.GetUserFromCtx(ctx)

	res, err := c.service.getTeammates(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// @Summary Get user played maps
// @Tags 	Analytics
// @Produce json
// @Param   body body 		GetPlayedMapsRequest true "Body"
// @Success 200 {object} 	GetPlayedMapsResponse
// @Router /analytics/users/maps [post]
func (c *controller) getPlayedMaps(ctx *gin.Context) {
	var req GetPlayedMapsRequest
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	res, err := c.service.getPlayedMaps(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// @Summary Get last seen users
// @Tags 	Analytics
// @Produce json
// @Param   body body 		GetLastSeenUsersRequest true "Body"
// @Success 200 {object} 	GetLastSeenUsersResponse
// @Router /analytics/users/lastseen [post]
func (c *controller) getLastSeenUsers(ctx *gin.Context) {
	user, _ := util.GetUserFromCtx(ctx)

	var req GetLastSeenUsersRequest
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	if req.UserId != user.UserId {
		ctx.String(http.StatusForbidden, "")
		return
	}

	res, err := c.service.getLastSeenUsers(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// @Summary Get last games with user
// @Tags 	Analytics
// @Produce json
// @Param   body body 		GetLastSessionsWithUserRequest true "Body"
// @Success 200 {object} 	GetLastSessionsWithUserResponse
// @Router /analytics/users/lastgameswithuser [post]
func (c *controller) getLastGamesWithUser(ctx *gin.Context) {
	user, _ := util.GetUserFromCtx(ctx)

	var req GetLastSessionsWithUserRequest
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	if req.UserId != user.UserId {
		ctx.String(http.StatusForbidden, "")
		return
	}

	res, err := c.service.getLastGamesWithUser(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}
