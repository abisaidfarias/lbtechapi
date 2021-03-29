package controllers

import (
	"fmt"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
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

// Create creates a case
func (c *deviceController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var device request.Device

		err := ctx.ShouldBindJSON(&device)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.deviceService.Create(&device)

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
		return
	}
}

// Get list all cases
func (c *deviceController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		devices, err := c.deviceService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, devices)
		return
	}

}

// Create creates a case
func (c *deviceController) GetById() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string

		id = ctx.Param("id")

		if id == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%w", utils.ErrorInvalidURLParams)})
			return
		}

		device, err := c.deviceService.GetById(id)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, device)
		return
	}
}

// Create creates a case
func (c *deviceController) Update() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var id string

		id = ctx.Param("id")

		var device request.Device

		err := ctx.ShouldBindJSON(&device)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.deviceService.Update(id, &device)

		if err != nil {
			if utils.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utils.ErrorDuplicated.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return

		}

		ctx.Status(http.StatusOK)
		return
	}
}

// Create creates a case
func (c *deviceController) Delete() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string

		id = ctx.Param("id")

		err := c.deviceService.Delete(id)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusOK)
		return

	}
}
