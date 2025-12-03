package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IBrandController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
}

// AuthController implementation of the interface
type brandController struct {
	brandService services.IBrandService
}

//NewUserController is the constructor
func NewBrandController(brandService services.IBrandService) IBrandController {
	return &brandController{
		brandService: brandService,
	}
}

// Create godoc
// @Summary Crear nueva marca
// @Description Crea una nueva marca en el sistema
// @Tags Brands
// @Accept json
// @Produce json
// @Security Bearer
// @Param brand body request.Brand true "Datos de la marca"
// @Success 201 {object} map[string]string "Marca creada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 409 {object} map[string]string "Marca duplicada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /brand [post]
func (c *brandController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var brand request.Brand

		err := ctx.ShouldBindJSON(&brand)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := c.brandService.Create(&brand)

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
// @Summary Listar marcas
// @Description Obtiene la lista de todas las marcas
// @Tags Brands
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.Brand "Lista de marcas"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /brand [get]
func (c *brandController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		brands, err := c.brandService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, brands)
	}

}
