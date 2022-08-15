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

// Create creates a category
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
func (c *deviceTrackingController) AddTrakingLog() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var trackingLog *request.TrackingLogMultiple
		userID := ctx.MustGet("userID").(string)

		err := ctx.ShouldBindJSON(&trackingLog)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.deviceTrackingService.AddTrakingLog(trackingLog, userID)

		if err != nil {
			if utils.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utils.ErrorDuplicated.Error()})
				return
			}
			ctx.Status(http.StatusInternalServerError)
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusOK)
	}
}
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
