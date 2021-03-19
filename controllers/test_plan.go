package controllers

import (
	"fmt"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"

	"github.com/gin-gonic/gin"
)

// ITestPlanController controller
type ITestPlanController interface {
	Create() gin.HandlerFunc
	GetById() gin.HandlerFunc
	Get() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
}

// testPlanController implementation of the interface
type testPlanController struct {
	testPlanService services.ITestPlanService
}

//NewUserController is the constructor
func NewTestPlanController(testPlanService services.ITestPlanService) ITestPlanController {
	return &testPlanController{
		testPlanService: testPlanService,
	}
}

// Create creates a Plan
func (c *testPlanController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var testPlan request.TestPlan

		err := ctx.ShouldBindJSON(&testPlan)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.testPlanService.Create(&testPlan)

		if err != nil {
			if utils.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utils.ErrorDuplicatedTestCode.Error()})
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

// Get list all Plans
func (c *testPlanController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		testPlans, err := c.testPlanService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, testPlans)
		return
	}

}

// Create creates a Plan
func (c *testPlanController) GetById() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string

		id = ctx.Param("id")

		if id == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%w", utils.ErrorInvalidURLParams)})
			return
		}

		testPlan, err := c.testPlanService.GetById(id)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, testPlan)
		return
	}
}

// Create creates a Plan
func (c *testPlanController) Update() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var id string

		id = ctx.Param("id")

		var testPlan request.TestPlan

		err := ctx.ShouldBindJSON(&testPlan)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.testPlanService.Update(id, &testPlan)

		if err != nil {
			if utils.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utils.ErrorDuplicatedTestCode.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return

		}

		ctx.Status(http.StatusOK)
		return
	}
}

// Create creates a Plan
func (c *testPlanController) Delete() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string

		id = ctx.Param("id")

		err := c.testPlanService.Delete(id)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusOK)
		return
	}
}
