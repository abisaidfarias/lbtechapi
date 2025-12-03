package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type ICountryController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
	Delete() gin.HandlerFunc
	Update() gin.HandlerFunc
}

// AuthController implementation of the interface
type countryController struct {
	countryService services.ICountryService
}

//NewUserController is the constructor
func NewCountryController(countryService services.ICountryService) ICountryController {
	return &countryController{
		countryService: countryService,
	}
}

// Create godoc
// @Summary Crear nuevo país
// @Description Crea un nuevo país en el sistema
// @Tags Countries
// @Accept json
// @Produce json
// @Security Bearer
// @Param country body request.Country true "Datos del país"
// @Success 201 "País creado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 409 {object} map[string]string "País duplicado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /country [post]
func (c *countryController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var country request.Country

		err := ctx.ShouldBindJSON(&country)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.countryService.Create(&country)

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
// @Summary Listar países
// @Description Obtiene la lista de todos los países
// @Tags Countries
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.Country "Lista de países"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /country [get]
func (c *countryController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		countries, err := c.countryService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, countries)
	}

}

// Delete godoc
// @Summary Eliminar país
// @Description Elimina un país del sistema
// @Tags Countries
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del país"
// @Success 200 "País eliminado exitosamente"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "País no encontrado"
// @Failure 409 "No se puede eliminar, existen homologaciones asociadas"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /country/{id} [delete]
func (c *countryController) Delete() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string = ctx.Param("id")

		hasHomologations, err := c.countryService.Delete(id)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}
		if hasHomologations {
			ctx.Status(http.StatusConflict)
			return
		}
		ctx.Status(http.StatusOK)
	}
}

// Update godoc
// @Summary Actualizar país
// @Description Actualiza un país existente
// @Tags Countries
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del país"
// @Param country body request.Country true "Datos actualizados del país"
// @Success 200 "País actualizado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "País no encontrado"
// @Failure 409 {object} map[string]string "Conflicto con datos existentes"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /country/{id} [put]
func (c *countryController) Update() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var id string = ctx.Param("id")

		var country request.Country

		err := ctx.ShouldBindJSON(&country)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.countryService.Update(id, &country)

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