package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IHomologationController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
	GetReport() gin.HandlerFunc
	GetCategoriesWithTest() gin.HandlerFunc
	UpdateTestResult() gin.HandlerFunc
	PhaseChange() gin.HandlerFunc
	GetHomologationFails() gin.HandlerFunc
	CreateFailTest() gin.HandlerFunc
	UpdateDocument() gin.HandlerFunc
	Delete() gin.HandlerFunc
	Update() gin.HandlerFunc
	ExportHomologation() gin.HandlerFunc
	ExportFailTest() gin.HandlerFunc
	UpdateFailTest() gin.HandlerFunc
}

// AuthController implementation of the interface
type homologationController struct {
	homologationService services.IHomologationService
}

//NewUserController is the constructor
func NewHomologationController(homologationService services.IHomologationService) IHomologationController {
	return &homologationController{
		homologationService: homologationService,
	}
}

// Create godoc
// @Summary Crear nueva homologación
// @Description Crea una nueva homologación en el sistema
// @Tags Homologations
// @Accept json
// @Produce json
// @Security Bearer
// @Param homologation body request.Homologation true "Datos de la homologación"
// @Success 201 "Homologación creada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 409 {object} map[string]interface{} "Conflicto - error con código"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /homologation [post]
func (c *homologationController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var homologation request.Homologation
		userID := ctx.MustGet("userID").(string)
		err := ctx.ShouldBindJSON(&homologation)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		customError, err := c.homologationService.Create(&homologation, userID)
		if customError != nil {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": customError.Err,
				"code":  customError.Code,
			})
			return
		}

		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusCreated)
	}
}

// Get godoc
// @Summary Listar homologaciones
// @Description Obtiene la lista de todas las homologaciones
// @Tags Homologations
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.Homologation "Lista de homologaciones"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /homologation [get]
func (c *homologationController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userID := ctx.MustGet("userID").(string)
		homologations, err := c.homologationService.Get(userID)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, homologations)
	}

}

// GetReport godoc
// @Summary Obtener reporte de homologación
// @Description Obtiene el reporte completo de una homologación
// @Tags Homologations
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID de la homologación"
// @Success 200 {object} responses.HomologationReport "Reporte de la homologación"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Homologación no encontrada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /homologation/{id}/report [get]
func (c *homologationController) GetReport() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id := ctx.Param("id")
		homologationReport, err := c.homologationService.GetReport(id)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, homologationReport)
	}
}

// GetCategoriesWithTest godoc
// @Summary Obtener categorías con tests
// @Description Obtiene las categorías de prueba con sus tests para una homologación
// @Tags Homologations
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID de la homologación"
// @Success 200 {array} responses.TestCategoryExpanded "Categorías con tests"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Homologación no encontrada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /homologation/{id}/categories/test [get]
func (c *homologationController) GetCategoriesWithTest() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id := ctx.Param("id")
		homologationReport, err := c.homologationService.GetCategoriesWithTest(id)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, homologationReport)
	}

}

// UpdateTestResult godoc
// @Summary Actualizar resultado de prueba
// @Description Actualiza el resultado de una prueba en la homologación
// @Tags Homologations
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID de la homologación"
// @Param testResult body request.TestResultResume true "Resultado de la prueba"
// @Success 200 "Resultado actualizado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Homologación no encontrada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /homologation/{id} [put]
func (c *homologationController) UpdateTestResult() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id := ctx.Param("id")

		var testResult request.TestResultResume

		err := ctx.ShouldBindJSON(&testResult)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = c.homologationService.UpdateTestResult(id, testResult)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusOK)
	}

}

// PhaseChange godoc
// @Summary Cambiar fase de homologación
// @Description Cambia la fase/estado de una homologación
// @Tags Homologations
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID de la homologación"
// @Param homologation body request.HomologationResume true "Datos de la fase"
// @Success 200 "Fase actualizada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Homologación no encontrada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /homologation/{id}/phase [put]
func (c *homologationController) PhaseChange() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id := ctx.Param("id")
		userID := ctx.MustGet("userID").(string)
		var homologation *request.HomologationResume

		err := ctx.ShouldBindJSON(&homologation)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = c.homologationService.PhaseChange(id, homologation, userID)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusOK)
	}

}

// GetHomologationFails godoc
// @Summary Obtener pruebas fallidas
// @Description Obtiene las pruebas fallidas de una homologación
// @Tags Homologations
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID de la homologación"
// @Success 200 {array} responses.TestResult "Lista de pruebas fallidas"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Homologación no encontrada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /homologation/{id}/test/fails [get]
func (c *homologationController) GetHomologationFails() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id := ctx.Param("id")

		testResults, err := c.homologationService.GetHomologationFails(id)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, testResults)
	}

}

// CreateFailTest godoc
// @Summary Crear prueba fallida
// @Description Registra una nueva prueba fallida en la homologación
// @Tags Homologations
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID de la homologación"
// @Param testResult body request.TestResultResume true "Datos de la prueba fallida"
// @Success 200 "Prueba fallida registrada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Homologación no encontrada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /homologation/{id}/failTest [post]
func (c *homologationController) CreateFailTest() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id := ctx.Param("id")

		var testResult *request.TestResultResume
		userID := ctx.MustGet("userID").(string)
		err := ctx.ShouldBindJSON(&testResult)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = c.homologationService.CreateFailTestResult(id, testResult,userID)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusOK)
	}

}

// UpdateDocument godoc
// @Summary Actualizar documento de homologación
// @Description Actualiza los documentos asociados a una homologación
// @Tags Homologations
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID de la homologación"
// @Param homologation body request.Homologation true "Datos de los documentos"
// @Success 200 "Documentos actualizados exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Homologación no encontrada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /homologation/{id}/document [put]
func (c *homologationController) UpdateDocument() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id := ctx.Param("id")

		var homologation *request.Homologation

		err := ctx.ShouldBindJSON(&homologation)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = c.homologationService.UpdateDocument(id, homologation)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusOK)
	}

}

// Delete godoc
// @Summary Eliminar homologación
// @Description Elimina una homologación del sistema
// @Tags Homologations
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID de la homologación"
// @Success 200 "Homologación eliminada exitosamente"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Homologación no encontrada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /homologation/{id} [delete]
func (c *homologationController) Delete() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string = ctx.Param("id")

		err := c.homologationService.Delete(id)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}
		ctx.Status(http.StatusOK)
	}
}

// Update godoc
// @Summary Actualizar homologación
// @Description Actualiza los datos de una homologación existente
// @Tags Homologations
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID de la homologación"
// @Param homologation body request.Homologation true "Datos actualizados de la homologación"
// @Success 200 "Homologación actualizada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Homologación no encontrada"
// @Failure 409 {object} map[string]string "Conflicto con datos existentes"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /homologation/{id}/homologation [put]
func (c *homologationController) Update() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var id string = ctx.Param("id")

		var homologation request.Homologation

		err := ctx.ShouldBindJSON(&homologation)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.homologationService.Update(id, &homologation)

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

// ExportHomologation godoc
// @Summary Exportar homologaciones a Excel
// @Description Exporta todas las homologaciones del usuario a un archivo Excel
// @Tags Homologations
// @Accept json
// @Produce application/octet-stream
// @Security Bearer
// @Success 200 {file} binary "Archivo Excel"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /export/homologation [get]
func (c *homologationController) ExportHomologation() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)
		file, err := c.homologationService.ExportHomologation(userID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		downloadName := fmt.Sprintf("%s%s", time.Now().UTC().Format("01-02-2006 15:04:05"), ".xlsx")
		ctx.Header("Content-Description", "File Transfer")
		ctx.Header("Content-Disposition", "attachment; filename="+downloadName)
		ctx.Header("Content-Type", "application/octet-stream")
		ctx.Header("Content-Transfer-Encoding", "binary")
		ctx.Data(http.StatusOK, "application/octet-stream", file.Bytes())
		ctx.Status(http.StatusOK)
	}
}

// ExportFailTest godoc
// @Summary Exportar pruebas fallidas a Excel
// @Description Exporta las pruebas fallidas de una homologación a un archivo Excel
// @Tags Homologations
// @Accept json
// @Produce application/octet-stream
// @Security Bearer
// @Param id path string true "ID de la homologación"
// @Success 200 {file} binary "Archivo Excel"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Homologación no encontrada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /export/fail-test/{id} [get]
func (c *homologationController) ExportFailTest() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string = ctx.Param("id")
		file, err := c.homologationService.ExportFailTest(id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		downloadName := fmt.Sprintf("%s%s", time.Now().UTC().Format("01-02-2006 15:04:05"), ".xlsx")
		ctx.Header("Content-Description", "File Transfer")
		ctx.Header("Content-Disposition", "attachment; filename="+downloadName)
		ctx.Header("Content-Type", "application/octet-stream")
		ctx.Header("Content-Transfer-Encoding", "binary")
		ctx.Data(http.StatusOK, "application/octet-stream", file.Bytes())
		ctx.Status(http.StatusOK)
	}
}

// UpdateFailTest godoc
// @Summary Actualizar pruebas fallidas
// @Description Actualiza las pruebas fallidas de una homologación
// @Tags Homologations
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID de la homologación"
// @Param testFails body request.TestFails true "Datos de las pruebas fallidas"
// @Success 200 "Pruebas fallidas actualizadas exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Homologación no encontrada"
// @Failure 409 {object} map[string]string "Conflicto con datos existentes"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /homologation/{id}/failTest [put]
func (c *homologationController) UpdateFailTest() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var id string = ctx.Param("id")

		var testFail request.TestFails

		err := ctx.ShouldBindJSON(&testFail)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.homologationService.UpdateFailTest(id, testFail.TestResults)

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
