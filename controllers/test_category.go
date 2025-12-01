package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type ITestCategoryController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
}

// AuthController implementation of the interface
type testCategoryController struct {
	testCategoryService services.ITestCategoryService
}

//NewUserController is the constructor
func NewTestCategoryController(testCategoryService services.ITestCategoryService) ITestCategoryController {
	return &testCategoryController{
		testCategoryService: testCategoryService,
	}
}

// Create godoc
// @Summary Crear nueva categoría de prueba
// @Description Crea una nueva categoría de prueba en el sistema
// @Tags Test Categories
// @Accept json
// @Produce json
// @Security Bearer
// @Param category body request.TestCategory true "Datos de la categoría"
// @Success 201 {object} map[string]string "Categoría creada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 409 {object} map[string]string "Categoría duplicada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /test-categories [post]
func (c *testCategoryController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var category request.TestCategory

		err := ctx.ShouldBindJSON(&category)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := c.testCategoryService.Create(&category)

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
// @Summary Listar categorías de prueba
// @Description Obtiene la lista de todas las categorías de prueba
// @Tags Test Categories
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.TestCategory "Lista de categorías"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /test-categories [get]
func (c *testCategoryController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		categories, err := c.testCategoryService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, categories)
	}

}
