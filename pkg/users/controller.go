package users

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type userController struct {
	service *UserService
}

// @Summary Creates a new user
// @Tags 	Users
// @Produce json
// @Param   user body    CreateUserRequest true "User JSON"
// @Success 201 {object} CreateUserResponse
// @Router /users [post]
func (c *userController) create(ctx *gin.Context) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	id, err := c.service.FindCreateFind(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, CreateUserResponse{
		Id: id,
	})
}

// @Summary Filter users
// @Tags 	Users
// @Produce json
// @Param   user body    FilterUsersRequest true "Filter JSON"
// @Success 201 {object} FilterUsersResponse
// @Router /users/filter [post]
func (c *userController) filter(ctx *gin.Context) {
	var req FilterUsersRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	res, err := c.service.filter(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// @Summary Get user with detailed by id
// @Tags 	Users
// @Produce json
// @Param   id path   	 	int true "User id"
// @Success 200 {object} 	FilterUsersResponseUser
// @Router /users/{id}/detailed [get]
func (c *userController) getUserDetailed(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	item, err := c.service.getUserDetailed(id)
	if err != nil {
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, item)
}
