package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IDashboardController interface {
	Get() gin.HandlerFunc
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

// Get list all categories
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
