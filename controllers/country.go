package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type ICountryController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
}

// AuthController implementation of the interface
type countryController struct {
	countryService services.ICountryService
}

//NewUserController is the constructor
func NewCountryController(countryService services.ICountryService) ICountryController {
	return &countryController{
		countryService: countryService,
	}
}

// Create creates a category
func (c *countryController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var country request.Country

		err := ctx.ShouldBindJSON(&country)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.countryService.Create(&country)

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

// Get list all categories
func (c *countryController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		countries, err := c.countryService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, countries)
		return
	}

}
