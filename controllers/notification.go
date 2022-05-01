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

// Create creates a category
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

// Get list all categories
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
