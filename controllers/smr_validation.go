package controllers

import (
	"errors"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

type ISmrValidationController interface {
	Get() gin.HandlerFunc
	Approve() gin.HandlerFunc
	Reject() gin.HandlerFunc
}

type smrValidationController struct {
	service services.ISmrValidationService
}

func NewSmrValidationController(service services.ISmrValidationService) ISmrValidationController {
	return &smrValidationController{service: service}
}

// Get godoc
// @Summary SMR without testing: validation data
// @Description Status, the last generated report of a request with the same device technical model, FM Radio of the device, stamp, stamps catalog and audit log. Internal users only.
// @Tags Multibanda
// @Produce json
// @Security Bearer
// @Param id path string true "multibanda id"
// @Success 200 {object} responses.SmrValidation
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /multibanda/{id}/smr-validation [get]
func (c *smrValidationController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := c.service.Get(ctx.Param("id"), ctx.MustGet("userID").(string))
		if err != nil {
			handleSmrValidationError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// Approve godoc
// @Summary SMR without testing: approve
// @Description Confirms FM Radio and the stamp; the request moves to Under evaluation. A comment is required when an existing FM Radio or stamp is corrected.
// @Tags Multibanda
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "multibanda id"
// @Param body body request.SmrValidationApprove true "decision"
// @Success 200 {object} responses.SmrValidation
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /multibanda/{id}/smr-validation/approve [post]
func (c *smrValidationController) Approve() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body request.SmrValidationApprove
		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := c.service.Approve(ctx.Param("id"), ctx.MustGet("userID").(string), &body)
		if err != nil {
			handleSmrValidationError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// Reject godoc
// @Summary SMR without testing: reject
// @Description Closes the request (phase Complete, status Rejected). comment is required.
// @Tags Multibanda
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "multibanda id"
// @Param body body request.SmrValidationReject true "reason"
// @Success 200 {object} responses.SmrValidation
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /multibanda/{id}/smr-validation/reject [post]
func (c *smrValidationController) Reject() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body request.SmrValidationReject
		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := c.service.Reject(ctx.Param("id"), ctx.MustGet("userID").(string), &body)
		if err != nil {
			handleSmrValidationError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func handleSmrValidationError(ctx *gin.Context, err error) {
	switch {
	case utils.IsValidationError(err):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, utils.ErrorForbidden):
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		handleErrorResponse(ctx, err)
	}
}
