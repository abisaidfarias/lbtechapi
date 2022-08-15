package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type ILocationController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
}

// AuthController implementation of the interface
type locationController struct {
	locationService services.ILocationService
}

//NewUserController is the constructor
func NewLocationController(locationService services.ILocationService) ILocationController {
	return &locationController{
		locationService: locationService,
	}
}

// Create creates a category
func (c *locationController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var location request.Location

		err := ctx.ShouldBindJSON(&location)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := c.locationService.Create(&location)

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
func (c *locationController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		locations, err := c.locationService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, locations)
	}

}
