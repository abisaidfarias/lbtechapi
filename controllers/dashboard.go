package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IDashboardController interface {
	Get() gin.HandlerFunc
	GetGeneralInfo() gin.HandlerFunc
}

// AuthController implementation of the interface
type dashboardController struct {
	dashboardService services.IDashboardService
}

//NewUserController is the constructor
func NewDashboardController(dashboardService services.IDashboardService) IDashboardController {
	return &dashboardController{
		dashboardService: dashboardService,
	}
}

// Get godoc
// @Summary Obtener dashboard
// @Description Obtiene los datos del dashboard para el usuario
// @Tags Dashboard
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} responses.DashboardReport "Datos del dashboard"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /dashboard [get]
func (c *dashboardController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userID := ctx.MustGet("userID").(string)
		dashboards, err := c.dashboardService.Get(userID)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, dashboards)
	}

}

// GetGeneralInfo godoc
// @Summary Obtener información general del dashboard
// @Description Obtiene la información general del dashboard
// @Tags Dashboard
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} responses.DashboardInfo "Información general"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /dashboard/info [get]
func (c *dashboardController) GetGeneralInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userID := ctx.MustGet("userID").(string)
		dashboards, err := c.dashboardService.GetGeneralInfo(userID)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, dashboards)
	}

}
