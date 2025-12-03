package controllers

import (
	"fmt"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// ITestCaseController controller
type ITestCaseController interface {
	Create() gin.HandlerFunc
	GetById() gin.HandlerFunc
	Get() gin.HandlerFunc
	Update() gin.HandlerFunc
	Upgrade() gin.HandlerFunc
	Delete() gin.HandlerFunc
	FileUpload() gin.HandlerFunc
}

// testCaseController implementation of the interface
type testCaseController struct {
	testCaseService services.ITestCaseService
}

//NewUserController is the constructor
func NewTestCaseController(testCaseService services.ITestCaseService) ITestCaseController {
	return &testCaseController{
		testCaseService: testCaseService,
	}
}

// Create godoc
// @Summary Crear nuevo caso de prueba
// @Description Crea un nuevo caso de prueba en el sistema
// @Tags Test Cases
// @Accept json
// @Produce json
// @Security Bearer
// @Param testCase body request.TestCase true "Datos del caso de prueba"
// @Success 201 "Caso de prueba creado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 409 {object} map[string]string "Caso de prueba duplicado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /test-cases [post]
func (c *testCaseController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var testCase request.TestCase

		err := ctx.ShouldBindJSON(&testCase)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.testCaseService.Create(&testCase)

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
// @Summary Listar casos de prueba
// @Description Obtiene la lista de todos los casos de prueba
// @Tags Test Cases
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.TestCase "Lista de casos de prueba"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /test-cases [get]
func (c *testCaseController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		testCases, err := c.testCaseService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, testCases)
	}

}

// GetById godoc
// @Summary Obtener caso de prueba por ID
// @Description Obtiene un caso de prueba específico por su ID
// @Tags Test Cases
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del caso de prueba"
// @Success 200 {object} responses.TestCase "Caso de prueba encontrado"
// @Failure 400 {object} map[string]string "Parámetros inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Caso de prueba no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /test-cases/{id} [get]
func (c *testCaseController) GetById() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string = ctx.Param("id")

		if id == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%w", utils.ErrorInvalidURLParams)})
			return
		}

		testCase, err := c.testCaseService.GetById(id)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, testCase)
	}
}

// Update godoc
// @Summary Actualizar caso de prueba
// @Description Actualiza un caso de prueba existente
// @Tags Test Cases
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del caso de prueba"
// @Param testCase body request.TestCase true "Datos actualizados del caso"
// @Success 200 "Caso actualizado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Caso no encontrado"
// @Failure 409 {object} map[string]string "Conflicto con datos existentes"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /test-cases/{id} [put]
func (c *testCaseController) Update() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var id string = ctx.Param("id")

		var testCase request.TestCase

		err := ctx.ShouldBindJSON(&testCase)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.testCaseService.Update(id, &testCase)

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

// Upgrade godoc
// @Summary Actualizar versión de caso de prueba
// @Description Actualiza la versión de un caso de prueba
// @Tags Test Cases
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del caso de prueba"
// @Param testCase body request.TestCase true "Datos de la nueva versión"
// @Success 200 "Versión actualizada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Caso no encontrado"
// @Failure 409 {object} map[string]string "Conflicto con datos existentes"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /test-cases/{id}/upgrade [put]
func (c *testCaseController) Upgrade() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var id string = ctx.Param("id")

		var testCase request.TestCase

		err := ctx.ShouldBindJSON(&testCase)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.testCaseService.Upgrade(id, &testCase)

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
// @Summary Eliminar caso de prueba
// @Description Elimina un caso de prueba del sistema
// @Tags Test Cases
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del caso de prueba"
// @Success 200 "Caso eliminado exitosamente"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Caso no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /test-cases/{id} [delete]
func (c *testCaseController) Delete() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string = ctx.Param("id")

		err := c.testCaseService.Delete(id)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusOK)
	}
}

// FileUpload godoc
// @Summary Cargar archivo de casos de prueba
// @Description Procesa y carga casos de prueba desde un archivo
// @Tags Test Cases
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param file formData file true "Archivo de casos de prueba"
// @Success 201 {object} responses.TestCaseFileUpload "Casos cargados exitosamente"
// @Failure 400 {object} map[string]string "Formato de archivo inválido"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /test-cases/fileUpload [post]
func (c *testCaseController) FileUpload() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fileHeader, err := ctx.FormFile("file")
		if err != nil {
			ctx.String(400, "file format error")
			return
		}
		res, err := c.testCaseService.ProcessFile(fileHeader)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, res)
	}
}
