package controllers

import (
	"fmt"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"

	"github.com/gin-gonic/gin"
)

// IProfileController controller
type IProfileController interface {
	Create() gin.HandlerFunc
	GetById() gin.HandlerFunc
	Get() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
}

// profileController implementation of the interface
type profileController struct {
	profileService services.IProfileService
}

//NewUserController is the constructor
func NewProfileController(profileService services.IProfileService) IProfileController {
	return &profileController{
		profileService: profileService,
	}
}

// Create creates a case
func (c *profileController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var profile request.Profile

		err := ctx.ShouldBindJSON(&profile)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.profileService.Create(&profile)

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
func (c *profileController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		profiles, err := c.profileService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, profiles)
		return
	}

}

// Create creates a case
func (c *profileController) GetById() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string

		id = ctx.Param("id")

		if id == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%w", utils.ErrorInvalidURLParams)})
			return
		}

		profile, err := c.profileService.GetById(id)

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, profile)
		return
	}
}

// Create creates a case
func (c *profileController) Update() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var id string

		id = ctx.Param("id")

		var profile request.Profile

		err := ctx.ShouldBindJSON(&profile)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.profileService.Update(id, &profile)

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
func (c *profileController) Delete() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string

		id = ctx.Param("id")

		err, canDelete := c.profileService.Delete(id)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}
		if !canDelete {
			ctx.Status(http.StatusConflict)
			return
		}

		ctx.Status(http.StatusOK)
		return

	}
}
