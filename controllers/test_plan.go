package controllers

import (
	"fmt"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
)

// ITestPlanController controller
type ITestPlanController interface {
	Create() gin.HandlerFunc
	GetById() gin.HandlerFunc
	Get() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
}

// testPlanController implementation of the interface
type testPlanController struct {
	testPlanService services.ITestPlanService
}

//NewUserController is the constructor
func NewTestPlanController(testPlanService services.ITestPlanService) ITestPlanController {
	return &testPlanController{
		testPlanService: testPlanService,
	}
}

// Create godoc
// @Summary Crear nuevo plan de pruebas
// @Description Crea un nuevo plan de pruebas en el sistema
// @Tags Test Plans
// @Accept json
// @Produce json
// @Security Bearer
// @Param testPlan body request.TestPlan true "Datos del plan de pruebas"
// @Success 201 "Plan de pruebas creado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 409 {object} map[string]string "Plan duplicado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /test-plan [post]
func (c *testPlanController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var testPlanRequest request.TestPlan
		err := ctx.ShouldBindJSON(&testPlanRequest)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userID := ctx.MustGet("userID").(string)
		oid, _ := primitive.ObjectIDFromHex(userID)
		testPlanRequest.UserID = oid
		err = c.testPlanService.Create(&testPlanRequest)
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
// @Summary Listar planes de pruebas
// @Description Obtiene la lista de todos los planes de pruebas
// @Tags Test Plans
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.TestPlan "Lista de planes de pruebas"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /test-plan [get]
func (c *testPlanController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		testPlans, err := c.testPlanService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, testPlans)
	}

}

// GetById godoc
// @Summary Obtener plan de pruebas por ID
// @Description Obtiene un plan de pruebas específico por su ID
// @Tags Test Plans
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del plan de pruebas"
// @Success 200 {object} responses.TestPlan "Plan de pruebas encontrado"
// @Failure 400 {object} map[string]string "Parámetros inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Plan no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /test-plan/{id} [get]
func (c *testPlanController) GetById() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string = ctx.Param("id")

		if id == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%w", utils.ErrorInvalidURLParams)})
			return
		}
		testPlan, err := c.testPlanService.GetById(id)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, testPlan)
	}
}

// Update godoc
// @Summary Actualizar plan de pruebas
// @Description Actualiza un plan de pruebas existente
// @Tags Test Plans
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del plan de pruebas"
// @Param testPlan body request.TestPlan true "Datos actualizados del plan"
// @Success 200 "Plan actualizado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Plan no encontrado"
// @Failure 409 {object} map[string]string "Conflicto con datos existentes"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /test-plan/{id} [put]
func (c *testPlanController) Update() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var id string = ctx.Param("id")

		var testPlan request.TestPlan

		err := ctx.ShouldBindJSON(&testPlan)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.testPlanService.Update(id, &testPlan)

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
// @Summary Eliminar plan de pruebas
// @Description Elimina un plan de pruebas del sistema
// @Tags Test Plans
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del plan de pruebas"
// @Success 200 "Plan eliminado exitosamente"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Plan no encontrado"
// @Failure 409 "No se puede eliminar, existen homologaciones asociadas"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /test-plan/{id} [delete]
func (c *testPlanController) Delete() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string = ctx.Param("id")

		hasHomologations, err := c.testPlanService.Delete(id)
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
