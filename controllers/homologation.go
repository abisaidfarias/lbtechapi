package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IHomologationController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
}

// AuthController implementation of the interface
type homologationController struct {
	homologationService services.IHomologationService
}

//NewUserController is the constructor
func NewHomologationController(homologationService services.IHomologationService) IHomologationController {
	return &homologationController{
		homologationService: homologationService,
	}
}

// Create creates a category
func (c *homologationController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var homologation request.Homologation

		err := ctx.ShouldBindJSON(&homologation)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		customError, err := c.homologationService.Create(&homologation)
		if customError != nil {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": customError.Err,
				"code":  customError.Code,
			})
			return
		}

		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			handleErrorResponse(ctx, err)
			return
		}

		ctx.Status(http.StatusCreated)
		return
	}
}

func (c *homologationController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userID := ctx.MustGet("userID").(string)
		homologations, err := c.homologationService.Get(userID)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, homologations)
		return
	}

}
