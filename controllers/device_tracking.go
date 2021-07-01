package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IDeviceTrackingController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
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

// Create creates a category
func (c *deviceTrackingController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var deviceTracking request.DeviceTracking

		err := ctx.ShouldBindJSON(&deviceTracking)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.deviceTrackingService.Create(&deviceTracking)

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

// Get list all categories
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
