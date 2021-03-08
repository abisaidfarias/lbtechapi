package controllers

import (
	"errors"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
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
			case errors.Is(err, utils.ErrorInvalidCredentials):
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
