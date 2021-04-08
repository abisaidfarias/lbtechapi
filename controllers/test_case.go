package controllers

import (
	"fmt"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// ITestCaseController controller
type ITestCaseController interface {
	Create() gin.HandlerFunc
	GetById() gin.HandlerFunc
	Get() gin.HandlerFunc
	Update() gin.HandlerFunc
	Upgrade() gin.HandlerFunc
	Delete() gin.HandlerFunc
	FileUpload() gin.HandlerFunc
}

// testCaseController implementation of the interface
type testCaseController struct {
	testCaseService services.ITestCaseService
}

//NewUserController is the constructor
func NewTestCaseController(testCaseService services.ITestCaseService) ITestCaseController {
	return &testCaseController{
		testCaseService: testCaseService,
	}
}

// Create creates a case
func (c *testCaseController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var testCase request.TestCase

		err := ctx.ShouldBindJSON(&testCase)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.testCaseService.Create(&testCase)

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
func (c *testCaseController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		testCases, err := c.testCaseService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, testCases)
		return
	}

}

// Create creates a case
func (c *testCaseController) GetById() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string

		id = ctx.Param("id")

		if id == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%w", utils.ErrorInvalidURLParams)})
			return
		}

		testCase, err := c.testCaseService.GetById(id)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, testCase)
		return
	}
}

// Create creates a case
func (c *testCaseController) Update() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var id string

		id = ctx.Param("id")

		var testCase request.TestCase

		err := ctx.ShouldBindJSON(&testCase)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.testCaseService.Update(id, &testCase)

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
func (c *testCaseController) Upgrade() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var id string

		id = ctx.Param("id")

		var testCase request.TestCase

		err := ctx.ShouldBindJSON(&testCase)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.testCaseService.Upgrade(id, &testCase)

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
func (c *testCaseController) Delete() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string

		id = ctx.Param("id")

		err := c.testCaseService.Delete(id)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusOK)
		return
	}
}

func (c *testCaseController) FileUpload() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fileHeader, err := ctx.FormFile("file")
		if err != nil {
			ctx.String(400, "file format error")
			return
		}
		res, err := c.testCaseService.ProcessFile(fileHeader)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, res)
	}
}
