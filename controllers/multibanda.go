package controllers

import (
	"errors"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

type IMultibandaController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
	PhaseChange() gin.HandlerFunc
}

type multibandaController struct {
	multibandaService services.IMultibandaService
}

func NewMultibandaController(multibandaService services.IMultibandaService) IMultibandaController {
	return &multibandaController{
		multibandaService: multibandaService,
	}
}

// Create godoc
// @Summary Crear nuevo registro Multibanda
// @Description Crea un registro Multibanda en fase Planning (current_phase = 0)
// @Tags Multibanda
// @Accept json
// @Produce json
// @Security Bearer
// @Param multibanda body request.Multibanda true "Datos del registro Multibanda"
// @Success 201 {object} map[string]string "Registro creado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin permiso para crear Multibanda"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /multibanda [post]
func (c *multibandaController) Create() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var multibanda request.Multibanda
		userID := ctx.MustGet("userID").(string)

		if err := ctx.ShouldBindJSON(&multibanda); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := c.multibandaService.Create(&multibanda, userID)
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
// @Summary Listar registros Multibanda
// @Description Obtiene la lista de registros Multibanda según el alcance del usuario
// @Tags Multibanda
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.MultibandaExpanded "Lista de registros Multibanda"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin permiso para leer Multibanda"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /multibanda [get]
func (c *multibandaController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)

		multibandas, err := c.multibandaService.Get(userID)
		if err != nil {
			if errors.Is(err, utils.ErrorForbidden) {
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, multibandas)
	}
}

// PhaseChange godoc
// @Summary Cambiar fase de Multibanda
// @Description Cambia la fase/estado de un registro Multibanda
// @Tags Multibanda
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del registro Multibanda"
// @Param multibanda body request.MultibandaResume true "Datos de la fase"
// @Success 200 "Fase actualizada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin permiso para escribir Multibanda"
// @Failure 404 {object} map[string]string "Registro Multibanda no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /multibanda/{id}/phase [put]
func (c *multibandaController) PhaseChange() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		userID := ctx.MustGet("userID").(string)
		var multibanda request.MultibandaResume

		if err := ctx.ShouldBindJSON(&multibanda); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := c.multibandaService.PhaseChange(id, &multibanda, userID); err != nil {
			if errors.Is(err, utils.ErrorForbidden) {
				ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusOK)
	}
}
