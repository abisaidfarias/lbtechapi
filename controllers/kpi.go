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

// GetVolumeChart godoc
// @Summary Obtener gráfica de volumen
// @Description Obtiene los datos de la gráfica de volumen por rango de fechas
// @Tags KPI
// @Accept json
// @Produce json
// @Security Bearer
// @Param start path string true "Fecha inicio (RFC3339)"
// @Param end path string true "Fecha fin (RFC3339)"
// @Success 200 {object} responses.VolumeChart "Datos de la gráfica"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 422 {object} map[string]string "Formato de fecha inválido"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /kpi/volume/{start}/{end} [get]
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

// GetTimeChart godoc
// @Summary Obtener gráfica de tiempos
// @Description Obtiene los datos de la gráfica de tiempos por rango de fechas
// @Tags KPI
// @Accept json
// @Produce json
// @Security Bearer
// @Param start path string true "Fecha inicio (RFC3339)"
// @Param end path string true "Fecha fin (RFC3339)"
// @Success 200 {object} responses.TimeChart "Datos de la gráfica"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 422 {object} map[string]string "Formato de fecha inválido"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /kpi/time/{start}/{end} [get]
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
