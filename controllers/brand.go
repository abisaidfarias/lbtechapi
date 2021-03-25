package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IBrandController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
}

// AuthController implementation of the interface
type brandController struct {
	brandService services.IBrandService
}

//NewUserController is the constructor
func NewBrandController(brandService services.IBrandService) IBrandController {
	return &brandController{
		brandService: brandService,
	}
}

// Create creates a category
func (c *brandController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var brand request.Brand

		err := ctx.ShouldBindJSON(&brand)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.brandService.Create(&brand)

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
func (c *brandController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		brands, err := c.brandService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, brands)
		return
	}

}
