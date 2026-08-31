package controllers

import (
	"errors"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

type IMultibandaReportController interface {
	GetForm() gin.HandlerFunc
	SaveDraft() gin.HandlerFunc
	Generate() gin.HandlerFunc
	GetStampImage() gin.HandlerFunc
}

type multibandaReportController struct {
	multibandaReportService services.IMultibandaReportService
}

func NewMultibandaReportController(s services.IMultibandaReportService) IMultibandaReportController {
	return &multibandaReportController{multibandaReportService: s}
}

// GetForm godoc
// @Summary Obtener el formulario del reporte automático Multi-banda
// @Description Devuelve los datos precargados del dispositivo, el borrador guardado (si existe), el alcance según el tipo de proceso y los catálogos necesarios para renderizar las matrices de resultados.
// @Tags MultibandaReport
// @Produce json
// @Security Bearer
// @Param id path string true "ID del proceso Multi-banda"
// @Success 200 {object} responses.MultibandaReportForm
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin acceso al proceso"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /multibanda/{id}/report [get]
func (c *multibandaReportController) GetForm() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)
		form, err := c.multibandaReportService.GetForm(ctx.Param("id"), userID)
		if err != nil {
			handleMultibandaReportError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, form)
	}
}

// SaveDraft godoc
// @Summary Guardar borrador del reporte automático Multi-banda
// @Description Guarda el reporte de forma incremental sin exigir que esté completo, para que el ingeniero pueda retomarlo después.
// @Tags MultibandaReport
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del proceso Multi-banda"
// @Param body body request.MultibandaReportSave true "Datos del reporte"
// @Success 200 {object} responses.MultibandaReportForm
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin acceso al proceso"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /multibanda/{id}/report/draft [put]
func (c *multibandaReportController) SaveDraft() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)

		var body request.MultibandaReportSave
		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		form, err := c.multibandaReportService.SaveDraft(ctx.Param("id"), &body, userID)
		if err != nil {
			handleMultibandaReportError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, form)
	}
}

// Generate godoc
// @Summary Generar el PDF del reporte automático Multi-banda
// @Description Valida que todos los campos y evidencias obligatorios estén completos, genera el PDF y lo almacena en el proceso, reemplazando el PDF generado anteriormente.
// @Tags MultibandaReport
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del proceso Multi-banda"
// @Param body body request.MultibandaReportSave true "Datos del reporte"
// @Success 200 {object} responses.MultibandaReportGenerated
// @Failure 400 {object} map[string]string "Faltan campos o evidencias obligatorias"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin acceso al proceso"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /multibanda/{id}/report/generate [post]
func (c *multibandaReportController) Generate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)

		var body request.MultibandaReportSave
		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := c.multibandaReportService.Generate(ctx.Param("id"), &body, userID)
		if err != nil {
			handleMultibandaReportError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// GetStampImage godoc
// @Summary Obtener la imagen de un sello del catálogo
// @Description Devuelve el PNG del sello indicado para que el selector de sellos pueda mostrarlo. Las imágenes están embebidas en el binario.
// @Tags MultibandaReport
// @Produce png
// @Security Bearer
// @Param code path string true "Código del sello (ver catálogo del formulario)"
// @Success 200 {file} file "Imagen PNG del sello"
// @Failure 404 {object} map[string]string "Sello no encontrado"
// @Router /multibanda/report/stamps/{code}/image [get]
func (c *multibandaReportController) GetStampImage() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		stamp, ok := enums.StampByCode(ctx.Param("code"))
		if !ok {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "stamp not found"})
			return
		}

		image := utils.MultibandaStampImage(stamp.ImageKey)
		if len(image) == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "stamp image not available"})
			return
		}

		// The catalog is fixed and the assets ship with the binary, so this is
		// safe to cache hard; a new build is what changes them.
		ctx.Header("Cache-Control", "public, max-age=86400")
		ctx.Data(http.StatusOK, "image/png", image)
	}
}

func handleMultibandaReportError(ctx *gin.Context, err error) {
	switch {
	case utils.IsValidationError(err):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, utils.ErrorForbidden):
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		handleErrorResponse(ctx, err)
	}
}
