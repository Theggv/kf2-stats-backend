package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type userController struct {
	service *UserService
}

// CreateUser godoc
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

	id, err := c.service.CreateUser(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, CreateUserResponse{
		Id: id,
	})
}
