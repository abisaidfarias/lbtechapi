package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
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

// SignIn godoc
// @Summary Iniciar sesión
// @Description Autentica un usuario y devuelve un token JWT
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body request.AuthCredentials true "Credenciales de usuario"
// @Success 200 {object} responses.AuthResponse "Usuario autenticado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "Credenciales incorrectas"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /sign-in [post]
func (c *authController) SignIn() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var credentials request.AuthCredentials

		err := ctx.ShouldBindJSON(&credentials)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userRes, err := c.authService.SignIn(&credentials)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, *userRes)
	}
}
