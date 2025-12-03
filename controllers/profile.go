package controllers

import (
	"fmt"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"

	"github.com/gin-gonic/gin"
)

// IProfileController controller
type IProfileController interface {
	Create() gin.HandlerFunc
	GetById() gin.HandlerFunc
	Get() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
}

// profileController implementation of the interface
type profileController struct {
	profileService services.IProfileService
}

//NewUserController is the constructor
func NewProfileController(profileService services.IProfileService) IProfileController {
	return &profileController{
		profileService: profileService,
	}
}

// Create godoc
// @Summary Crear nuevo perfil
// @Description Crea un nuevo perfil de usuario en el sistema
// @Tags Profiles
// @Accept json
// @Produce json
// @Security Bearer
// @Param profile body request.Profile true "Datos del perfil"
// @Success 201 "Perfil creado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 409 {object} map[string]string "Perfil duplicado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /profile [post]
func (c *profileController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var profile request.Profile

		err := ctx.ShouldBindJSON(&profile)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userID := ctx.MustGet("userID").(string)
		profile.UserID = userID
		err = c.profileService.Create(&profile)

		if err != nil {
			if utils.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utils.ErrorDuplicated.Error()})
				return
			}
			ctx.Status(http.StatusInternalServerError)
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusCreated)
	}
}

// Get godoc
// @Summary Listar perfiles
// @Description Obtiene la lista de todos los perfiles
// @Tags Profiles
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.Profile "Lista de perfiles"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /profile [get]
func (c *profileController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userID := ctx.MustGet("userID").(string)
		profiles, err := c.profileService.Get(userID)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, profiles)
	}

}

// GetById godoc
// @Summary Obtener perfil por ID
// @Description Obtiene un perfil específico por su ID
// @Tags Profiles
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del perfil"
// @Success 200 {object} responses.Profile "Perfil encontrado"
// @Failure 400 {object} map[string]string "Parámetros inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Perfil no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /profile/{id} [get]
func (c *profileController) GetById() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		id := ctx.Param("id")

		if id == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%w", utils.ErrorInvalidURLParams)})
			return
		}

		profile, err := c.profileService.GetById(id)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, profile)
	}
}

// Update godoc
// @Summary Actualizar perfil
// @Description Actualiza un perfil existente
// @Tags Profiles
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del perfil"
// @Param profile body request.Profile true "Datos actualizados del perfil"
// @Success 200 "Perfil actualizado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Perfil no encontrado"
// @Failure 409 {object} map[string]string "Conflicto con datos existentes"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /profile/{id} [put]
func (c *profileController) Update() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		id := ctx.Param("id")

		var profile request.Profile

		err := ctx.ShouldBindJSON(&profile)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.profileService.Update(id, &profile)

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

// Delete godoc
// @Summary Eliminar perfil
// @Description Elimina un perfil del sistema
// @Tags Profiles
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del perfil"
// @Success 200 "Perfil eliminado exitosamente"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Perfil no encontrado"
// @Failure 409 "No se puede eliminar, existen dependencias"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /profile/{id} [delete]
func (c *profileController) Delete() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		id := ctx.Param("id")

		canDelete, err := c.profileService.Delete(id)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}
		if !canDelete {
			ctx.Status(http.StatusConflict)
			return
		}

		ctx.Status(http.StatusOK)

	}
}
