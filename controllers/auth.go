package controllers

import (
	"errors"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"
	utilsErrors "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"github.com/gin-gonic/gin"
)

// IAuthController controller
type IAuthController interface {
	SignIn() gin.HandlerFunc
	ForgotPassword() gin.HandlerFunc
	SetPassword() gin.HandlerFunc
	ValidatePasswordToken() gin.HandlerFunc
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
// @Failure 403 {object} map[string]string "Contraseña pendiente de configuración"
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
			handleAuthErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, *userRes)
	}
}

func (c *authController) ForgotPassword() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body request.ForgotPassword
		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := c.authService.ForgotPassword(&body); err != nil {
			handleAuthErrorResponse(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, responses.MessageResponse{
			Message: "If an account exists for that email, password instructions were sent.",
		})
	}
}

func (c *authController) SetPassword() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body request.SetPassword
		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := c.authService.SetPassword(&body); err != nil {
			handleAuthErrorResponse(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, responses.MessageResponse{Message: "Password updated successfully"})
	}
}

func (c *authController) ValidatePasswordToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Query("token")
		resp, err := c.authService.ValidatePasswordToken(token)
		if err != nil {
			handleAuthErrorResponse(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, resp)
	}
}

func handleAuthErrorResponse(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, utilsErrors.ErrorInvalidCredentials):
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, utilsErrors.ErrorPasswordNotSet):
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "password_not_set"})
	case errors.Is(err, utilsErrors.ErrorInvalidPasswordToken):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case utils.IsValidationError(err):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
