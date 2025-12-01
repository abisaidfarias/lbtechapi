package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type INotificationController interface {
	Create() gin.HandlerFunc
	GetByCompany() gin.HandlerFunc
}

// AuthController implementation of the interface
type notificationController struct {
	notificationService services.INotificationService
}

//NewUserController is the constructor
func NewNotificationController(notificationService services.INotificationService) INotificationController {
	return &notificationController{
		notificationService: notificationService,
	}
}

// Create godoc
// @Summary Crear nueva notificación
// @Description Crea una nueva notificación en el sistema
// @Tags Notifications
// @Accept json
// @Produce json
// @Security Bearer
// @Param notification body request.Notification true "Datos de la notificación"
// @Success 201 {object} map[string]string "Notificación creada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 409 {object} map[string]string "Notificación duplicada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /notification [post]
func (c *notificationController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var notification request.Notification

		err := ctx.ShouldBindJSON(&notification)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := c.notificationService.Create(&notification)

		if err != nil {
			if utils.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utils.ErrorDuplicated.Error()})
				return
			}
			ctx.Status(http.StatusInternalServerError)
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"id": id})
	}
}

// GetByCompany godoc
// @Summary Obtener notificaciones por compañía
// @Description Obtiene las notificaciones asociadas a una compañía específica
// @Tags Notifications
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "ID de la compañía"
// @Success 200 {array} responses.Notification "Lista de notificaciones"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /notification/company/{id} [get]
func (c *notificationController) GetByCompany() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		notifications, err := c.notificationService.GetByCompany(id)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, notifications)
	}

}
