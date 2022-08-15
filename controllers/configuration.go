package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IConfigurationController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
}

// AuthController implementation of the interface
type configurationController struct {
	configurationService services.IConfigurationService
}

//NewUserController is the constructor
func NewConfigurationController(configurationService services.IConfigurationService) IConfigurationController {
	return &configurationController{
		configurationService: configurationService,
	}
}

// Create creates a category
func (c *configurationController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var configuration request.Configuration

		err := ctx.ShouldBindJSON(&configuration)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := c.configurationService.Create(&configuration)

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
func (c *configurationController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		configurations, err := c.configurationService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, configurations)
	}

}
