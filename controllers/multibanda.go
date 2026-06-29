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

type IMultibandaController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
	Update() gin.HandlerFunc
	PhaseChange() gin.HandlerFunc
	Delete() gin.HandlerFunc
	PatchRequestDelete() gin.HandlerFunc
	RejectRequestDelete() gin.HandlerFunc
	ExportMultibanda() gin.HandlerFunc
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
// @Success 200 {array} responses.MultibandaExpanded "Lista de registros Multibanda (incluye request_delete)"
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

// Update godoc
// @Summary Actualizar registro Multibanda
// @Description Actualiza todos los campos editables del registro Multibanda. Solo usuarios internos.
// @Tags Multibanda
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del registro Multibanda"
// @Param multibanda body request.Multibanda true "Datos actualizados del registro Multibanda"
// @Success 200 "Registro actualizado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Solo usuarios internos"
// @Failure 404 {object} map[string]string "Registro Multibanda no encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /multibanda/{id} [put]
func (c *multibandaController) Update() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		userID := ctx.MustGet("userID").(string)
		var multibanda request.Multibanda

		if err := ctx.ShouldBindJSON(&multibanda); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := c.multibandaService.Update(id, &multibanda, userID); err != nil {
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

// Delete godoc
// @Summary Eliminar registro Multibanda
// @Description Solo usuario interno. Elimina el registro de forma definitiva.
// @Tags Multibanda
// @Produce json
// @Security Bearer
// @Param id path string true "ID del registro Multibanda"
// @Success 200 {object} responses.DeleteProcessResult "Resultado de la eliminación"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin permiso"
// @Failure 404 {object} map[string]string "No encontrado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /multibanda/{id} [delete]
func (c *multibandaController) Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		userID := ctx.MustGet("userID").(string)

		result, err := c.multibandaService.Delete(id, userID)
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
// @Summary Actualizar solicitud de eliminación Multibanda
// @Description Externo: solicitar borrado con request_delete true.
// @Tags Multibanda
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID del registro Multibanda"
// @Param body body request.RequestDeletePatch true "request_delete true"
// @Success 200 {object} responses.DeleteProcessResult "Resultado"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin permiso"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /multibanda/{id}/request-delete [patch]
func (c *multibandaController) PatchRequestDelete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		userID := ctx.MustGet("userID").(string)
		var body request.RequestDeletePatch

		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := c.multibandaService.PatchRequestDelete(id, &body, userID)
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
// @Summary Rechazar solicitud de eliminación Multibanda
// @Description Usuario interno: cancela una solicitud de borrado pendiente (request_delete pasa a false).
// @Tags Multibanda
// @Produce json
// @Security Bearer
// @Param id path string true "ID del registro Multibanda"
// @Success 200 {object} responses.DeleteProcessResult "Resultado"
// @Failure 400 {object} map[string]string "Sin solicitud pendiente o ID inválido"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Solo usuarios internos"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /multibanda/{id}/reject-delete [patch]
func (c *multibandaController) RejectRequestDelete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		userID := ctx.MustGet("userID").(string)

		result, err := c.multibandaService.RejectRequestDelete(id, userID)
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

// ExportMultibanda godoc
// @Summary Exportar registros Multibanda a Excel
// @Description Exporta todos los registros Multibanda visibles para el usuario a un archivo Excel
// @Tags Multibanda
// @Produce application/octet-stream
// @Security Bearer
// @Success 200 {file} binary "Archivo Excel"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 403 {object} map[string]string "Sin permiso"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /export/multibanda [get]
func (c *multibandaController) ExportMultibanda() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)
		file, err := c.multibandaService.ExportMultibanda(userID)
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
