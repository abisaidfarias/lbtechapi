package controllers

import (
	"net/http"
	"time"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IKpiController interface {
	GetVolumeChart() gin.HandlerFunc
	GetTimeChart() gin.HandlerFunc
}

// AuthController implementation of the interface
type kpiController struct {
	kpiService services.IKpiService
}

//NewUserController is the constructor
func NewKpiController(kpiService services.IKpiService) IKpiController {
	return &kpiController{
		kpiService: kpiService,
	}
}

// Get list all categories
func (c *kpiController) GetVolumeChart() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userID := ctx.MustGet("userID").(string)
		startDate, err := time.Parse(time.RFC3339, ctx.Param("start"))
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		endDate, err := time.Parse(time.RFC3339, ctx.Param("end"))
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		volumeCharts, err := c.kpiService.GetVolumeChart(userID, startDate, endDate)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, volumeCharts)
	}

}
func (c *kpiController) GetTimeChart() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userID := ctx.MustGet("userID").(string)
		startDate, err := time.Parse(time.RFC3339, ctx.Param("start"))
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		endDate, err := time.Parse(time.RFC3339, ctx.Param("end"))
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		timesChart, err := c.kpiService.GetTimeChart(userID, startDate, endDate)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, timesChart)
	}

}
