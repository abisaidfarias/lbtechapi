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
type IDeviceTrackingController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
	AddTrakingLog() gin.HandlerFunc
	Delete() gin.HandlerFunc
	Update() gin.HandlerFunc
	AdvancedSearch() gin.HandlerFunc
	AdvancedSearchOptions() gin.HandlerFunc
	ExportDeviceTracking() gin.HandlerFunc
}

// AuthController implementation of the interface
type deviceTrackingController struct {
	deviceTrackingService services.IDeviceTrackingService
}

//NewUserController is the constructor
func NewDeviceTrackingController(deviceTrackingService services.IDeviceTrackingService) IDeviceTrackingController {
	return &deviceTrackingController{
		deviceTrackingService: deviceTrackingService,
	}
}

// Create godoc
// @Summary Crear nuevo tracking de dispositivo
// @Description Crea un nuevo registro de tracking para un dispositivo
// @Tags Device Tracking
// @Accept json
// @Produce json
// @Security Bearer
// @Param deviceTracking body request.DeviceTracking true "Datos del tracking"
// @Success 201 "Tracking creado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 409 {object} map[string]string "Tracking duplicado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /device-tracking [post]
func (c *deviceTrackingController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var deviceTracking request.DeviceTracking

		userID := ctx.MustGet("userID").(string)
		err := ctx.ShouldBindJSON(&deviceTracking)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.deviceTrackingService.Create(&deviceTracking, userID)

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
// @Summary Listar trackings de dispositivos
// @Description Obtiene la lista de todos los trackings de dispositivos
// @Tags Device Tracking
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.DeviceTracking "Lista de trackings"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /device-tracking [get]
func (c *deviceTrackingController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)
		deviceTrackings, err := c.deviceTrackingService.Get(userID)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, deviceTrackings)
	}

}

// AddTrakingLog godoc
// @Summary Agregar log de tracking
// @Description Agrega un nuevo log a un tracking existente
// @Tags Device Tracking
// @Accept json
// @Produce json
// @Security Bearer
// @Param trackingLog body request.TrackingLogMultiple true "Datos del log"
// @Success 200 "Log agregado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 409 {object} map[string]string "Conflicto"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /device-tracking [put]
func (c *deviceTrackingController) AddTrakingLog() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var trackingLog *request.TrackingLogMultiple
		userID := ctx.MustGet("userID").(string)

		err := ctx.ShouldBindJSON(&trackingLog)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		trackingID, err := c.deviceTrackingService.AddTrakingLog(trackingLog, userID)

		if err != nil {
			if utils.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utils.ErrorDuplicated.Error()})
				return
			}
			ctx.Status(http.StatusInternalServerError)
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"tracking_id": trackingID})
	}
}

// Delete godoc
// @Summary Eliminar tracking
// @Description Elimina un tracking de dispositivo
// @Tags Device Tracking
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del tracking"
// @Success 200 "Tracking eliminado exitosamente"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Tracking no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /device-tracking/{id} [delete]
func (c *deviceTrackingController) Delete() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var ids string = ctx.Param("id")

		err := c.deviceTrackingService.Delete(ids)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}
		ctx.Status(http.StatusOK)
	}
}

// Update godoc
// @Summary Actualizar tracking
// @Description Actualiza un tracking de dispositivo existente
// @Tags Device Tracking
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del tracking"
// @Param deviceTracking body request.DeviceTrackingExpanded true "Datos actualizados"
// @Success 200 "Tracking actualizado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Tracking no encontrado"
// @Failure 409 {object} map[string]string "Conflicto con datos existentes"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /device-tracking/{id} [put]
func (c *deviceTrackingController) Update() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		id := ctx.Param("id")

		var deviceTracking request.DeviceTrackingExpanded

		err := ctx.ShouldBindJSON(&deviceTracking)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.deviceTrackingService.Update(id, &deviceTracking)

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

// AdvancedSearch godoc
// @Summary Búsqueda avanzada de trackings
// @Description Realiza una búsqueda avanzada de trackings con filtros
// @Tags Device Tracking
// @Accept json
// @Produce json
// @Security Bearer
// @Param searchOption body request.SearchOption true "Opciones de búsqueda"
// @Success 200 {array} responses.DeviceTracking "Resultados de la búsqueda"
// @Failure 400 {object} map[string]string "Parámetros de búsqueda inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /device-tracking/advanced-search [post]
func (c *deviceTrackingController) AdvancedSearch() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)
		var searchOption *request.SearchOption

		err := ctx.ShouldBindJSON(&searchOption)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		deviceTrackings, err := c.deviceTrackingService.AdvancedSearch(searchOption, userID)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, deviceTrackings)
	}

}

// AdvancedSearchOptions godoc
// @Summary Obtener opciones de búsqueda avanzada
// @Description Obtiene las opciones disponibles para la búsqueda avanzada
// @Tags Device Tracking
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} responses.SearchOption "Opciones de búsqueda"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /device-tracking/search-options [get]
func (c *deviceTrackingController) AdvancedSearchOptions() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)

		options, err := c.deviceTrackingService.AdvancedSearchOptions(userID)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, options)
	}

}

// ExportDeviceTracking godoc
// @Summary Exportar trackings a Excel
// @Description Exporta los trackings de dispositivos a un archivo Excel
// @Tags Device Tracking
// @Accept json
// @Produce application/octet-stream
// @Security Bearer
// @Param searchOption body request.SearchOption true "Opciones de búsqueda/filtro"
// @Success 200 {file} binary "Archivo Excel"
// @Failure 400 {object} map[string]string "Parámetros inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /export/device-tracking [post]
func (c *deviceTrackingController) ExportDeviceTracking() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)
		var searchOption *request.SearchOption

		err := ctx.ShouldBindJSON(&searchOption)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		file, err := c.deviceTrackingService.ExportDeviceTracking(searchOption, userID)
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
