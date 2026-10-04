package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"
	utilsErrors "github.com/abisaidfarias/lbtechapi/utils/errors"
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
	ChangePassword() gin.HandlerFunc
	ChangeMyPassword() gin.HandlerFunc
	Upgrade() gin.HandlerFunc
	GetByInternal() gin.HandlerFunc
	GetUserByCompany() gin.HandlerFunc
	ResendInvitation() gin.HandlerFunc
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

// Create godoc
// @Summary Crear nuevo usuario
// @Description Crea un nuevo usuario en el sistema
// @Tags Users
// @Accept json
// @Produce json
// @Security Bearer
// @Param user body request.User true "Datos del usuario"
// @Success 201 {object} responses.User "Usuario creado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /users [post]
func (c *userController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var user request.User

		err := ctx.ShouldBindJSON(&user)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userID := ctx.MustGet("userID").(string)
		user.UserID = userID
		err = c.userService.Create(&user)

		if err != nil {
			if utilsErrors.IsDuplicateUserEmailConflict(err) {
				respondDuplicateUserEmail(ctx)
				return
			}
			if errors.Is(err, utilsErrors.ErrorUserCreatedInviteFailed) {
				ctx.JSON(http.StatusCreated, gin.H{
					"message": "User created but invitation email could not be sent. Check FRONTEND_BASE_URL and email settings, then resend the invite.",
					"warning": err.Error(),
				})
				return
			}
			if utils.IsValidationError(err) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{
			"message": "User created. An email was sent to set their password.",
		})
	}
}

// Upgrade godoc
// @Summary Crear el primer usuario del sistema (solo una vez)
// @Description Este endpoint permite crear el primer usuario administrador del sistema. Solo puede ser utilizado si no existen usuarios previamente. Crea automáticamente un perfil "Admin" y una compañía "Default Company" si no existen.
// @Tags Users
// @Accept json
// @Produce json
// @Param user body request.User true "Datos del usuario"
// @Success 201 {object} map[string]string "Usuario creado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 403 {object} map[string]string "No permitido - ya existen usuarios"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /upgrade [post]
func (c *userController) Upgrade() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var user request.User

		err := ctx.ShouldBindJSON(&user)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		err = c.userService.Upgrade(&user)

		if err != nil {
			if utilsErrors.IsDuplicateUserEmailConflict(err) {
				respondDuplicateUserEmail(ctx)
				return
			}
			// Verificar si es el error de upgrade no permitido
			if err.Error() == utilsErrors.ErrorUpgradeNotAllowed.Error() {
				ctx.JSON(http.StatusForbidden, gin.H{"error": utilsErrors.ErrorUpgradeNotAllowed.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusCreated)
	}
}

func (c *userController) GetByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		if id == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%w", utilsErrors.ErrorInvalidURLParams)})
			return
		}

		user, err := c.userService.GetByID(id)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, *user)
	}

}
func (c *userController) GetByEmail() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var id string

		email := ctx.Param("email")

		if id == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%w", utilsErrors.ErrorInvalidURLParams)})
			return
		}

		user, err := c.userService.GetByEmail(email)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, *user)
	}

}

func (c *userController) Update() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id := ctx.Param("id")

		var user models.User

		err := ctx.ShouldBindJSON(&user)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.userService.Update(id, &user)

		if err != nil {
			if utilsErrors.IsDuplicateUserEmailConflict(err) {
				respondDuplicateUserEmail(ctx)
				return
			}
			handleErrorResponse(ctx, err)
			return

		}

		ctx.Status(http.StatusOK)

	}

}

// ResendInvitation godoc
// @Summary Resend user invitation email
// @Description Sends a new password-invite link to a user who has not set their password yet
// @Tags Users
// @Produce json
// @Security Bearer
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 429 {object} map[string]string
// @Router /users/{id}/resend-invitation [post]
func (c *userController) ResendInvitation() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		targetID := ctx.Param("id")
		if targetID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": utilsErrors.ErrorInvalidURLParams.Error()})
			return
		}
		actorID := ctx.MustGet("userID").(string)
		err := c.userService.ResendInvitation(actorID, targetID)
		if err != nil {
			if errors.Is(err, utilsErrors.ErrorUserNotPendingInvite) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if errors.Is(err, utilsErrors.ErrorInviteResendRateLimited) {
				ctx.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
				return
			}
			if errors.Is(err, utilsErrors.ErrorResourceNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			if errors.Is(err, utilsErrors.ErrorUserCreatedInviteFailed) {
				ctx.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Invitation email sent.",
		})
	}
}

func (c *userController) Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id := ctx.Param("id")

		err := c.userService.Delete(id)

		if err != nil {
			if errors.Is(err, utilsErrors.ErrorResourceNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusOK)

	}

}

// Get godoc
// @Summary Listar usuarios
// @Description Obtiene la lista de usuarios del sistema
// @Tags Users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.User "Lista de usuarios"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /users [get]
func (c *userController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)
		users, err := c.userService.Get(userID)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, users)
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

	}

}
func (c *userController) ChangePassword() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusGone, gin.H{"error": utilsErrors.ErrorDeprecatedPasswordEndpoint.Error()})
	}
}

func (c *userController) ChangeMyPassword() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)
		var changePassword request.ChangePassword
		if err := ctx.ShouldBindJSON(&changePassword); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := c.userService.ChangeMyPassword(userID, changePassword); err != nil {
			if errors.Is(err, utilsErrors.ErrorInvalidCredentials) {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return
		}
		ctx.Status(http.StatusOK)
	}
}
func (c *userController) GetByInternal() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		users, err := c.userService.GetInternalUser()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, users)

	}

}
func (c *userController) GetUserByCompany() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		companyId := ctx.Param("id")

		users, err := c.userService.GetUsersByCompany(companyId)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, users)

	}

}
