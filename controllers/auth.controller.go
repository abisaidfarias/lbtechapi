package controllers

import (
	"errors"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/viewmodels"
	"github.com/gin-gonic/gin"
)

// IAuthController controller
type IAuthController interface {
	SignIn(c viewmodels.AuthCredentials) viewmodels.UserResponse
}

// AuthController implementation of the interface
type AuthController struct {
	AuthService services.AuthService
}

// SignIn signs the user in
func (c *AuthController) SignIn() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		credentials := viewmodels.AuthCredentials{}

		err := ctx.BindJSON(&credentials)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, err)
		}

		userRes, err := c.AuthService.SignIn(&credentials)

		if err != nil {
			switch {
			case errors.Is(err, models.ErrorInvalidCredentials):
				ctx.JSON(http.StatusUnauthorized, err.Error())
				return
			case errors.Is(err, models.ErrorUnableToSignToken):
				ctx.JSON(http.StatusInternalServerError, err.Error())
				return
			default:
				ctx.JSON(http.StatusInternalServerError, err.Error())
				return
			}
		}

		ctx.JSON(http.StatusOK, *userRes)
	}
}

func (c *AuthController) SignUp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO move into full register info
		credentials := viewmodels.AuthCredentials{}

		err := ctx.BindJSON(&credentials)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, err)
		}

		err = c.AuthService.SignUp(&credentials)

		if err != nil {
			switch {
			case errors.Is(err, models.ErrorInvalidCredentials):
				ctx.JSON(http.StatusUnauthorized, err.Error())
				return
			case errors.Is(err, models.ErrorUnableToSignToken):
				ctx.JSON(http.StatusInternalServerError, err.Error())
				return
			default:
				ctx.JSON(http.StatusInternalServerError, err.Error())
				return
			}
		}

		ctx.Status(http.StatusCreated)

	}
}
