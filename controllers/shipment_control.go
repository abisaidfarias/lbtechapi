package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

type IShipmentControlController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
	Update() gin.HandlerFunc
	GetAvailableMultibandas() gin.HandlerFunc
	PhaseChange() gin.HandlerFunc
	Delete() gin.HandlerFunc
	PatchRequestDelete() gin.HandlerFunc
	RejectRequestDelete() gin.HandlerFunc
	ExportShipmentControl() gin.HandlerFunc
}

type shipmentControlController struct {
	shipmentControlService services.IShipmentControlService
}

func NewShipmentControlController(shipmentControlService services.IShipmentControlService) IShipmentControlController {
	return &shipmentControlController{
		shipmentControlService: shipmentControlService,
	}
}

// Create godoc
// @Summary Crear Shipment Control
// @Description Crea un control de embarque en fase Planning. Usuario externo: country se asigna automáticamente a Chile. Usuario interno: country es requerido.
// @Tags ShipmentControl
// @Accept json
// @Produce json
// @Security Bearer
// @Param shipmentControl body request.ShipmentControl true "Datos del control de embarque"
// @Success 201 {object} map[string]string "Control creado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin permiso"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /shipment-control [post]
func (c *shipmentControlController) Create() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body request.ShipmentControl
		userID := ctx.MustGet("userID").(string)

		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := c.shipmentControlService.Create(&body, userID)
		if err != nil {
			switch {
			case utils.IsValidationError(err):
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			case errors.Is(err, utils.ErrorForbidden):
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			default:
				ctx.Status(http.StatusInternalServerError)
				handleErrorResponse(ctx, err)
				return
			}
		}

		ctx.JSON(http.StatusCreated, gin.H{"id": id})
	}
}

// Get godoc
// @Summary Listar controles de embarque
// @Description Obtiene controles de embarque. Usuario interno: todas las companies en user.clients (si clients está vacío, todas). Usuario externo: solo su company.
// @Tags ShipmentControl
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.ShipmentControlExpanded "Lista de controles de embarque (incluye request_delete)"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin permiso"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /shipment-control [get]
func (c *shipmentControlController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)

		items, err := c.shipmentControlService.Get(userID)
		if err != nil {
			switch {
			case utils.IsValidationError(err):
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			case errors.Is(err, utils.ErrorForbidden):
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			default:
				handleErrorResponse(ctx, err)
				return
			}
		}

		ctx.JSON(http.StatusOK, items)
	}
}

// Update godoc
// @Summary Actualizar Shipment Control
// @Description Actualiza todos los campos editables del control de embarque. Solo usuarios internos.
// @Tags ShipmentControl
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del Shipment Control"
// @Param shipmentControl body request.ShipmentControl true "Datos actualizados del control de embarque"
// @Success 200 "Control actualizado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Solo usuarios internos"
// @Failure 404 {object} map[string]string "Control no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /shipment-control/{id} [put]
func (c *shipmentControlController) Update() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		userID := ctx.MustGet("userID").(string)
		var body request.ShipmentControl

		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := c.shipmentControlService.Update(id, &body, userID); err != nil {
			switch {
			case utils.IsValidationError(err):
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			case errors.Is(err, utils.ErrorForbidden):
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			default:
				handleErrorResponse(ctx, err)
				return
			}
		}

		ctx.Status(http.StatusOK)
	}
}

// GetAvailableMultibandas godoc
// @Summary Multibandas disponibles para Shipment Control
// @Description Obtiene multibandas del cliente agrupadas por device y software_version para crear controles de embarque
// @Tags ShipmentControl
// @Accept json
// @Produce json
// @Security Bearer
// @Param company query string false "Company ID (requerido para usuarios internos)"
// @Success 200 {object} responses.ShipmentControlAvailableResponse "Opciones disponibles"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin permiso"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /shipment-control/available-multibandas [get]
func (c *shipmentControlController) GetAvailableMultibandas() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)
		company := ctx.Query("company")

		response, err := c.shipmentControlService.GetAvailableMultibandas(userID, company)
		if err != nil {
			switch {
			case utils.IsValidationError(err):
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			case errors.Is(err, utils.ErrorForbidden):
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			default:
				handleErrorResponse(ctx, err)
				return
			}
		}

		ctx.JSON(http.StatusOK, response)
	}
}

// PhaseChange godoc
// @Summary Cambiar fase de Shipment Control
// @Description Actualiza la fase y los hitos de un control de embarque
// @Tags ShipmentControl
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del Shipment Control"
// @Param shipmentControl body request.ShipmentControlResume true "Datos de la fase"
// @Success 200 "Fase actualizada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin permiso"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /shipment-control/{id}/phase [put]
func (c *shipmentControlController) PhaseChange() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		userID := ctx.MustGet("userID").(string)
		var body request.ShipmentControlResume

		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := c.shipmentControlService.PhaseChange(id, &body, userID); err != nil {
			switch {
			case utils.IsValidationError(err):
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			case errors.Is(err, utils.ErrorForbidden):
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			default:
				handleErrorResponse(ctx, err)
				return
			}
		}

		ctx.Status(http.StatusOK)
	}
}

// Delete godoc
// @Summary Eliminar Shipment Control
// @Description Solo usuario interno. Elimina el registro de forma definitiva.
// @Tags ShipmentControl
// @Produce json
// @Security Bearer
// @Param id path string true "ID del Shipment Control"
// @Success 200 {object} responses.DeleteProcessResult "Resultado de la eliminación"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin permiso"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /shipment-control/{id} [delete]
func (c *shipmentControlController) Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		userID := ctx.MustGet("userID").(string)

		result, err := c.shipmentControlService.Delete(id, userID)
		if err != nil {
			switch {
			case utils.IsValidationError(err):
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			case errors.Is(err, utils.ErrorForbidden):
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			default:
				handleErrorResponse(ctx, err)
				return
			}
		}

		ctx.JSON(http.StatusOK, result)
	}
}

// PatchRequestDelete godoc
// @Summary Actualizar solicitud de eliminación Shipment Control
// @Description Externo: solicitar borrado con request_delete true.
// @Tags ShipmentControl
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del Shipment Control"
// @Param body body request.RequestDeletePatch true "request_delete true"
// @Success 200 {object} responses.DeleteProcessResult "Resultado"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin permiso"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /shipment-control/{id}/request-delete [patch]
func (c *shipmentControlController) PatchRequestDelete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		userID := ctx.MustGet("userID").(string)
		var body request.RequestDeletePatch

		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := c.shipmentControlService.PatchRequestDelete(id, &body, userID)
		if err != nil {
			switch {
			case utils.IsValidationError(err):
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			case errors.Is(err, utils.ErrorForbidden):
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			default:
				handleErrorResponse(ctx, err)
				return
			}
		}

		ctx.JSON(http.StatusOK, result)
	}
}

// RejectRequestDelete godoc
// @Summary Rechazar solicitud de eliminación Shipment Control
// @Description Usuario interno: cancela una solicitud de borrado pendiente (request_delete pasa a false).
// @Tags ShipmentControl
// @Produce json
// @Security Bearer
// @Param id path string true "ID del Shipment Control"
// @Success 200 {object} responses.DeleteProcessResult "Resultado"
// @Failure 400 {object} map[string]string "Sin solicitud pendiente o ID inválido"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Solo usuarios internos"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /shipment-control/{id}/reject-delete [patch]
func (c *shipmentControlController) RejectRequestDelete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		userID := ctx.MustGet("userID").(string)

		result, err := c.shipmentControlService.RejectRequestDelete(id, userID)
		if err != nil {
			switch {
			case utils.IsValidationError(err):
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			case errors.Is(err, utils.ErrorForbidden):
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			default:
				handleErrorResponse(ctx, err)
				return
			}
		}

		ctx.JSON(http.StatusOK, result)
	}
}

// ExportShipmentControl godoc
// @Summary Exportar Shipment Control a Excel
// @Description Exporta todos los controles de embarque visibles para el usuario a un archivo Excel
// @Tags ShipmentControl
// @Produce application/octet-stream
// @Security Bearer
// @Success 200 {file} binary "Archivo Excel"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin permiso"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /export/shipment-control [get]
func (c *shipmentControlController) ExportShipmentControl() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)
		file, err := c.shipmentControlService.ExportShipmentControl(userID)
		if err != nil {
			if errors.Is(err, utils.ErrorForbidden) {
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		downloadName := fmt.Sprintf("%s%s", time.Now().UTC().Format("01-02-2006 15:04:05"), ".xlsx")
		ctx.Header("Content-Description", "File Transfer")
		ctx.Header("Content-Disposition", "attachment; filename="+downloadName)
		ctx.Header("Content-Type", "application/octet-stream")
		ctx.Header("Content-Transfer-Encoding", "binary")
		ctx.Data(http.StatusOK, "application/octet-stream", file.Bytes())
	}
}
