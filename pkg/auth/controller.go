package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/config"
	"github.com/theggv/kf2-stats-backend/pkg/common/steamapi"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
)

type authController struct {
	service *AuthService
}

// @Summary Login via Steam OpenID
// @Tags 	Auth
// @Produce json
// @Param   user body    steamapi.ValidateOpenIdRequest true "User JSON"
// @Success 201 {object} AuthResponse
// @Router /auth/login [post]
func (c *authController) login(ctx *gin.Context) {
	var req steamapi.ValidateOpenIdRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	res, err := c.service.Login(req)
	if err != nil {
		ctx.String(http.StatusUnauthorized, err.Error())
		return
	}

	err = util.SetCookies(ctx, res.RefreshToken, config.Instance.JwtRefreshExpiresIn)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, &AuthResponse{
		AccessToken: res.AccessToken,
	})
}

// @Summary Refresh access token
// @Tags 	Auth
// @Produce json
// @Success 201 {object} AuthResponse
// @Router /auth/refresh [post]
func (c *authController) refresh(ctx *gin.Context) {
	cookie, err := ctx.Request.Cookie("refreshToken")
	if err != nil || cookie.Value == "" {
		ctx.String(http.StatusUnauthorized, err.Error())
		return
	}

	res, err := c.service.Refresh(cookie.Value)
	if err != nil {
		ctx.String(http.StatusUnauthorized, err.Error())
		return
	}

	err = util.SetCookies(ctx, res.RefreshToken, config.Instance.JwtRefreshExpiresIn)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, &AuthResponse{
		AccessToken: res.AccessToken,
	})
}

// @Summary Logout
// @Tags 	Auth
// @Produce json
// @Success 201
// @Router /auth/logout [post]
func (c *authController) logout(ctx *gin.Context) {
	cookie, err := ctx.Request.Cookie("refreshToken")
	if err != nil || cookie.Value == "" {
		ctx.String(http.StatusUnauthorized, err.Error())
		return
	}

	err = c.service.Logout(cookie.Value)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, nil)
}
