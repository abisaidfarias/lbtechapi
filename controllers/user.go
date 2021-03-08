package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IUserController interface {
	Create() gin.HandlerFunc
	GetByOid() gin.HandlerFunc
}

// AuthController implementation of the interface
type userController struct {
	userService services.IUserService
}

//NewUserController is the constructor
func NewUserController(userService services.IUserService) IUserController {
	return &userController{
		userService: userService,
	}
}

// SignUp creates a new user
func (c *userController) Create() gin.HandlerFunc {
	fmt.Println("test")
	return func(ctx *gin.Context) {

		var user request.UserRequest

		err := ctx.ShouldBindJSON(&user)

		// TODO add here the correct error response from custom password validator
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.userService.Create(&user)

		if err != nil {
			switch {
			case errors.Is(err, utils.ErrorInvalidCredentials):
				ctx.JSON(http.StatusUnauthorized, err.Error())
				return
			default:
				ctx.JSON(http.StatusInternalServerError, err.Error())
				return
			}
		}

		ctx.Status(http.StatusCreated)

	}
}
func (c *userController) GetByOid() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var id string

		err := ctx.ShouldBindJSON(&id)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := c.userService.GetByOID(id)

		if err != nil {
			switch {
			case errors.Is(err, utils.ErrorInvalidCredentials):
				ctx.JSON(http.StatusUnauthorized, err.Error())
				return
			default:
				ctx.JSON(http.StatusInternalServerError, err.Error())
				return
			}
		}
		ctx.JSON(http.StatusOK, *user)
	}

}
