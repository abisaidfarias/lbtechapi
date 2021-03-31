package controllers

import (
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
	GetByEmail() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
	Get() gin.HandlerFunc
	GetProfileByID() gin.HandlerFunc
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
		userID := ctx.MustGet("userID").(string)
		user.UserID = userID
		err = c.userService.Create(&user)

		if err != nil {
			if utils.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utils.ErrorDuplicated.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusCreated)
		return
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

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, *user)
		return
	}

}
func (c *userController) GetByEmail() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var id string

		email := ctx.Param("email")

		if id == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%w", utils.ErrorInvalidURLParams)})
			return
		}

		user, err := c.userService.GetByEmail(email)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, *user)
		return
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

		if err != nil {
			if utils.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utils.ErrorDuplicated.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return

		}

		ctx.Status(http.StatusOK)
		return
	}

}

func (c *userController) Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var id string

		id = ctx.Param("id")

		err := c.userService.Delete(id)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusOK)
		return
	}

}
func (c *userController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)
		users, err := c.userService.Get(userID)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, users)
		return
	}

}
func (c *userController) GetProfileByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userID := ctx.MustGet("userID").(string)
		profile, err := c.userService.GetProfileByID(userID)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, profile)
		return
	}

}
