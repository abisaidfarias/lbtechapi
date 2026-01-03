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
	ChangePassword() gin.HandlerFunc
	Upgrade() gin.HandlerFunc
	GetByInternal() gin.HandlerFunc
	GetUserByCompany() gin.HandlerFunc
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
			if utils.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utils.ErrorDuplicated.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusCreated)
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
			if utils.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utils.ErrorDuplicated.Error()})
				return
			}
			// Verificar si es el error de upgrade no permitido
			if err.Error() == utils.ErrorUpgradeNotAllowed.Error() {
				ctx.JSON(http.StatusForbidden, gin.H{"error": utils.ErrorUpgradeNotAllowed.Error()})
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
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%w", utils.ErrorInvalidURLParams)})
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
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%w", utils.ErrorInvalidURLParams)})
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
			if utils.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utils.ErrorDuplicated.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return

		}

		ctx.Status(http.StatusOK)

	}

}

func (c *userController) Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id := ctx.Param("id")

		err := c.userService.Delete(id)

		if err != nil {
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

		email := ctx.Param("id")

		var changePassword request.ChangePassword

		err := ctx.ShouldBindJSON(&changePassword)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = c.userService.ChangePassword(email, changePassword)

		if err != nil {
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
