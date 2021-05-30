package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IKpiController interface {
	GetVolumeChart() gin.HandlerFunc
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
		kpis, err := c.kpiService.GetVolumeChart(userID)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, kpis)
	}

}
