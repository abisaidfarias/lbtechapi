package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IPersonController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
	Delete() gin.HandlerFunc
	Update() gin.HandlerFunc
}

// AuthController implementation of the interface
type personController struct {
	personService services.IPersonService
}

//NewUserController is the constructor
func NewPersonController(personService services.IPersonService) IPersonController {
	return &personController{
		personService: personService,
	}
}

// Create godoc
// @Summary Crear nueva persona
// @Description Crea una nueva persona en el sistema
// @Tags Persons
// @Accept json
// @Produce json
// @Security Bearer
// @Param person body request.Person true "Datos de la persona"
// @Success 201 {object} map[string]string "Persona creada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 409 {object} map[string]string "Persona duplicada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /person [post]
func (c *personController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var person request.Person

		err := ctx.ShouldBindJSON(&person)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := c.personService.Create(&person)

		if err != nil {
			if utils.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utils.ErrorDuplicated.Error()})
				return
			}
			ctx.Status(http.StatusInternalServerError)
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"id": id})

	}
}

// Get godoc
// @Summary Listar personas
// @Description Obtiene la lista de todas las personas
// @Tags Persons
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.Person "Lista de personas"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /person [get]
func (c *personController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		companies, err := c.personService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, companies)
	}

}

// Delete godoc
// @Summary Eliminar persona
// @Description Elimina una persona del sistema
// @Tags Persons
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID de la persona"
// @Success 200 "Persona eliminada exitosamente"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Persona no encontrada"
// @Failure 409 "No se puede eliminar, existen dependencias"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /person/{id} [delete]
func (c *personController) Delete() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string = ctx.Param("id")

		hasRelations, err := c.personService.Delete(id)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}
		if hasRelations {
			ctx.Status(http.StatusConflict)
			return
		}
		ctx.Status(http.StatusOK)
	}
}

// Update godoc
// @Summary Actualizar persona
// @Description Actualiza una persona existente
// @Tags Persons
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID de la persona"
// @Param person body request.Person true "Datos actualizados de la persona"
// @Success 200 "Persona actualizada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Persona no encontrada"
// @Failure 409 {object} map[string]string "Conflicto con datos existentes"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /person/{id} [put]
func (c *personController) Update() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var id string = ctx.Param("id")

		var person request.Person

		err := ctx.ShouldBindJSON(&person)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.personService.Update(id, &person)

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
