package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	apputils "github.com/abisaidfarias/lbtechapi/utils"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type ICompanyController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
	Delete() gin.HandlerFunc
	Update() gin.HandlerFunc
	ClientIDAvailable() gin.HandlerFunc
	GetUsers() gin.HandlerFunc
}

// ClientIDAvailable godoc
// @Summary Check whether a Client ID is free
// @Description Live validation for the company form. exclude_id is the company being edited.
// @Tags Companies
// @Produce json
// @Security Bearer
// @Param client_id query string true "Client ID"
// @Param exclude_id query string false "Company being edited"
// @Success 200 {object} responses.CompanyClientIDAvailability
// @Failure 400 {object} map[string]string
// @Router /company/client-id-available [get]
func (c *companyController) ClientIDAvailable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := c.companyService.ClientIDAvailable(ctx.Query("client_id"), ctx.Query("exclude_id"))
		if err != nil {
			if apputils.IsValidationError(err) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// GetUsers godoc
// @Summary Users of a company
// @Description Users whose company or clients include it, with their profile name.
// @Tags Companies
// @Produce json
// @Security Bearer
// @Param id path string true "Company ID"
// @Success 200 {array} responses.CompanyUser
// @Failure 400 {object} map[string]string
// @Router /company/{id}/users [get]
func (c *companyController) GetUsers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		users, err := c.companyService.GetUsers(ctx.Param("id"))
		if err != nil {
			if apputils.IsValidationError(err) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, users)
	}
}

// AuthController implementation of the interface
type companyController struct {
	companyService services.ICompanyService
}

//NewUserController is the constructor
func NewCompanyController(companyService services.ICompanyService) ICompanyController {
	return &companyController{
		companyService: companyService,
	}
}

// Create godoc
// @Summary Crear nueva compañía
// @Description Crea una nueva compañía en el sistema
// @Tags Companies
// @Accept json
// @Produce json
// @Security Bearer
// @Param company body request.Company true "Datos de la compañía"
// @Success 201 "Compañía creada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 409 {object} map[string]string "Compañía duplicada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /company [post]
func (c *companyController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var company request.Company

		err := ctx.ShouldBindJSON(&company)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.companyService.Create(&company)

		if err != nil {
			if apputils.IsValidationError(err) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
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
// @Summary Listar compañías
// @Description Obtiene la lista de todas las compañías
// @Tags Companies
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.CompanyListItem "Lista de compañías"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /company [get]
func (c *companyController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		companies, err := c.companyService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, companies)
	}

}

// Delete godoc
// @Summary Eliminar compañía
// @Description Elimina una compañía del sistema
// @Tags Companies
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID de la compañía"
// @Success 200 "Compañía eliminada exitosamente"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Compañía no encontrada"
// @Failure 409 "No se puede eliminar, existen dependencias"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /company/{id} [delete]
func (c *companyController) Delete() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string = ctx.Param("id")

		hasRelations, err := c.companyService.Delete(id)
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
// @Summary Actualizar compañía
// @Description Actualiza una compañía existente
// @Tags Companies
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID de la compañía"
// @Param company body request.Company true "Datos actualizados de la compañía"
// @Success 200 "Compañía actualizada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Compañía no encontrada"
// @Failure 409 {object} map[string]string "Conflicto con datos existentes"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /company/{id} [put]
func (c *companyController) Update() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var id string = ctx.Param("id")

		var company request.Company

		err := ctx.ShouldBindJSON(&company)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.companyService.Update(id, &company)

		if err != nil {
			if apputils.IsValidationError(err) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
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
