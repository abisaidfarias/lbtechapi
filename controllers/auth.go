package controllers

import (
	"errors"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	util "github.com/abisaidfarias/lbtechapi/util/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels"
	"github.com/gin-gonic/gin"
)

// IAuthController controller
type IAuthController interface {
	SignIn() gin.HandlerFunc
}

// AuthController implementation of the interface
type authController struct {
	authService services.IAuthService
}

// NewAuthController is the constructor
func NewAuthController(authService services.IAuthService) IAuthController {
	return &authController{
		authService: authService,
	}
}

// SignIn signs the user in
func (c *authController) SignIn() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var credentials viewmodels.AuthCredentials

		err := ctx.ShouldBindJSON(&credentials)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userRes, err := c.authService.SignIn(&credentials)

		if err != nil {
			switch {
			case errors.Is(err, util.ErrorInvalidCredentials):
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			default:
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		ctx.JSON(http.StatusOK, *userRes)
	}
}

// SignUp creates a new user
// func (c *AuthController) SignUp() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		// TODO move into full register view model
// 		var credentials viewmodels.AuthCredentials

// 		err := ctx.ShouldBindJSON(&credentials)

// 		if err != nil {
// 			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		err = c.AuthService.SignUp(&credentials)

// 		if err != nil {
// 			switch {
// 			case errors.Is(err, models.ErrorInvalidCredentials):
// 				ctx.JSON(http.StatusUnauthorized, err.Error())
// 				return
// 			default:
// 				ctx.JSON(http.StatusInternalServerError, err.Error())
// 				return
// 			}
// 		}

// 		ctx.Status(http.StatusCreated)

// 	}
// }
