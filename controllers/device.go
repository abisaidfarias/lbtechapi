package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"
	utilsErrors "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"

	"github.com/gin-gonic/gin"
)

// IDeviceController controller
type IDeviceController interface {
	Create() gin.HandlerFunc
	GetById() gin.HandlerFunc
	Get() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
}

// deviceController implementation of the interface
type deviceController struct {
	deviceService services.IDeviceService
}

//NewUserController is the constructor
func NewDeviceController(deviceService services.IDeviceService) IDeviceController {
	return &deviceController{
		deviceService: deviceService,
	}
}

// Create godoc
// @Summary Crear nuevo dispositivo
// @Description Crea un nuevo dispositivo en el sistema
// @Tags Devices
// @Accept json
// @Produce json
// @Security Bearer
// @Param device body request.Device true "Datos del dispositivo"
// @Success 201 {object} responses.Device "Dispositivo creado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 409 {object} map[string]string "Dispositivo duplicado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /device [post]
func (c *deviceController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var device request.Device
		userID := ctx.MustGet("userID").(string)
		err := ctx.ShouldBindJSON(&device)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		deviceResponse, err := c.deviceService.Create(&device, userID)

		if err != nil {
			if utilsErrors.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utilsErrors.ErrorDuplicated.Error()})
				return
			}
			ctx.Status(http.StatusInternalServerError)
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, *deviceResponse)
	}
}

// Get godoc
// @Summary Listar dispositivos
// @Description Obtiene la lista de todos los dispositivos
// @Tags Devices
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.Device "Lista de dispositivos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /device [get]
func (c *deviceController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)

		devices, err := c.deviceService.Get(userID)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, devices)
	}

}

// GetById godoc
// @Summary Obtener dispositivo por ID
// @Description Obtiene un dispositivo específico por su ID
// @Tags Devices
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del dispositivo"
// @Success 200 {object} responses.Device "Dispositivo encontrado"
// @Failure 400 {object} map[string]string "Parámetros inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Dispositivo no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /device/{id} [get]
func (c *deviceController) GetById() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		id := ctx.Param("id")
		userID := ctx.MustGet("userID").(string)

		if id == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%w", utilsErrors.ErrorInvalidURLParams)})
			return
		}

		device, err := c.deviceService.GetById(id, userID)

		if err != nil {
			if errors.Is(err, utils.ErrorForbidden) {
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, device)
	}
}

// Update godoc
// @Summary Actualizar dispositivo
// @Description Actualiza un dispositivo existente
// @Tags Devices
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del dispositivo"
// @Param device body request.Device true "Datos actualizados del dispositivo"
// @Success 200 "Dispositivo actualizado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Dispositivo no encontrado"
// @Failure 409 {object} map[string]string "Conflicto con datos existentes"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /device/{id} [put]
func (c *deviceController) Update() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		id := ctx.Param("id")
		userID := ctx.MustGet("userID").(string)
		var device request.Device

		err := ctx.ShouldBindJSON(&device)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.deviceService.Update(id, &device, userID)

		if err != nil {
			if utilsErrors.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utilsErrors.ErrorDuplicated.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return

		}

		ctx.Status(http.StatusOK)
	}
}

// Delete godoc
// @Summary Eliminar dispositivo
// @Description Elimina un dispositivo del sistema
// @Tags Devices
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del dispositivo"
// @Success 200 "Dispositivo eliminado exitosamente"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 404 {object} map[string]string "Dispositivo no encontrado"
// @Failure 409 "No se puede eliminar, existen dependencias"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /device/{id} [delete]
func (c *deviceController) Delete() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string = ctx.Param("id")

		hasRelations, err := c.deviceService.Delete(id)
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
