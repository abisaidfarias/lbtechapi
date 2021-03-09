package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IUserController interface {
	Create() gin.HandlerFunc
	GetByID() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
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

	return func(ctx *gin.Context) {

		var user request.UserRequest

		err := ctx.ShouldBindJSON(&user)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.userService.Create(&user)

		if err != nil {
			switch {
			case errors.Is(err, utils.ErrorDuplicated):
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			default:
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		ctx.Status(http.StatusCreated)

	}
}

func (c *userController) GetByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var id string

		id = ctx.Param("id")

		if id == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%w", utils.ErrorInvalidURLParams)})
			return
		}

		user, err := c.userService.GetByID(id)

		// TODO check if switch is needed, which errors should we have here, think about general error handler
		if err != nil {
			switch {
			default:
				ctx.JSON(http.StatusInternalServerError, err.Error())
				return
			}
		}
		ctx.JSON(http.StatusOK, *user)
	}

}

func (c *userController) Update() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var id string

		id = ctx.Param("id")

		var user models.User

		err := ctx.ShouldBindJSON(&user)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.userService.Update(id, &user)

		// TODO check if switch is needed, which errors should we have here, think about general error handler
		if err != nil {
			switch {
			default:
				ctx.JSON(http.StatusInternalServerError, err.Error())
				return
			}
		}

		ctx.Status(http.StatusNoContent)
	}

}

func (c *userController) Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var id string

		id = ctx.Param("id")

		err := c.userService.Delete(id)

		// TODO check if switch is needed, which errors should we have here, think about general error handler
		if err != nil {
			switch {
			default:
				ctx.JSON(http.StatusInternalServerError, err.Error())
				return
			}
		}

		ctx.Status(http.StatusNoContent)
	}

}
