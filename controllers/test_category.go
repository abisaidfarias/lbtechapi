package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type ITestCategoryController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
}

// AuthController implementation of the interface
type testCategoryController struct {
	testCategoryService services.ITestCategoryService
}

//NewUserController is the constructor
func NewTestCategoryController(testCategoryService services.ITestCategoryService) ITestCategoryController {
	return &testCategoryController{
		testCategoryService: testCategoryService,
	}
}

// Create creates a category
func (c *testCategoryController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var category request.TestCategory

		err := ctx.ShouldBindJSON(&category)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := c.testCategoryService.Create(&category)

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
		return
	}
}

// Get list all categories
func (c *testCategoryController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		categories, err := c.testCategoryService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, categories)
		return
	}

}
