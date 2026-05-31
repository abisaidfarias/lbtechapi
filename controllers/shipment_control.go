package controllers

import (
	"errors"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

type IShipmentControlController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
	GetAvailableMultibandas() gin.HandlerFunc
	PhaseChange() gin.HandlerFunc
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
// @Success 200 {array} responses.ShipmentControlExpanded "Lista de controles de embarque"
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
