package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type ICompanyController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
}

// AuthController implementation of the interface
type companyController struct {
	companyService services.ICompanyService
}

//NewUserController is the constructor
func NewCompanyController(companyService services.ICompanyService) ICompanyController {
	return &companyController{
		companyService: companyService,
	}
}

// Create creates a category
func (c *companyController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var company request.Company

		err := ctx.ShouldBindJSON(&company)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.companyService.Create(&company)

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
func (c *companyController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		companies, err := c.companyService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, companies)
		return
	}

}
