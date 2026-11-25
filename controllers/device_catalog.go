package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

type IDeviceCatalogController interface {
	List() gin.HandlerFunc
	CreateValue() gin.HandlerFunc
	UpdateValue() gin.HandlerFunc
	SetOrder() gin.HandlerFunc
	DeleteValue() gin.HandlerFunc
}

type deviceCatalogController struct {
	service services.IDeviceCatalogService
}

func NewDeviceCatalogController(service services.IDeviceCatalogService) IDeviceCatalogController {
	return &deviceCatalogController{service: service}
}

// List godoc
// @Summary Device catalogs
// @Description Every catalog (device_type, rom, ram, memory_capability, usb) with its values in order and how many devices use each value.
// @Tags DeviceCatalogs
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.DeviceCatalog
// @Router /device-catalogs [get]
func (c *deviceCatalogController) List() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := c.service.List()
		if err != nil {
			handleDeviceCatalogError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// CreateValue godoc
// @Summary Add a catalog value
// @Description Added at the end of the order, active. Unique per category (case-insensitive).
// @Tags DeviceCatalogs
// @Accept json
// @Produce json
// @Security Bearer
// @Param category path string true "device_type | rom | ram | memory_capability | usb"
// @Param value body request.DeviceCatalogValueCreate true "value"
// @Success 201 {object} responses.DeviceCatalogValue
// @Failure 400 {object} map[string]string
// @Router /device-catalogs/{category}/values [post]
func (c *deviceCatalogController) CreateValue() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body request.DeviceCatalogValueCreate
		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := c.service.CreateValue(ctx.Param("category"), &body)
		if err != nil {
			handleDeviceCatalogError(ctx, err)
			return
		}
		ctx.JSON(http.StatusCreated, result)
	}
}

// UpdateValue godoc
// @Summary Update a catalog value
// @Description value, unit and active are optional. Renaming does not rewrite devices.
// @Tags DeviceCatalogs
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "value id"
// @Param value body request.DeviceCatalogValueUpdate true "changes"
// @Success 200 {object} responses.DeviceCatalogValue
// @Failure 400 {object} map[string]string
// @Router /device-catalogs/values/{id} [put]
func (c *deviceCatalogController) UpdateValue() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body request.DeviceCatalogValueUpdate
		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := c.service.UpdateValue(ctx.Param("id"), &body)
		if err != nil {
			handleDeviceCatalogError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// SetOrder godoc
// @Summary Reorder a catalog
// @Description ids must list every value of the category once; order becomes 0..n.
// @Tags DeviceCatalogs
// @Accept json
// @Produce json
// @Security Bearer
// @Param category path string true "category"
// @Param order body request.DeviceCatalogOrder true "ids in the new order"
// @Success 200 {array} responses.DeviceCatalogValue
// @Failure 400 {object} map[string]string
// @Router /device-catalogs/{category}/order [put]
func (c *deviceCatalogController) SetOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body request.DeviceCatalogOrder
		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := c.service.SetOrder(ctx.Param("category"), &body)
		if err != nil {
			handleDeviceCatalogError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// DeleteValue godoc
// @Summary Delete an unused catalog value
// @Description Only values used by no device; used ones must be deactivated.
// @Tags DeviceCatalogs
// @Security Bearer
// @Param id path string true "value id"
// @Success 204
// @Failure 400 {object} map[string]string
// @Router /device-catalogs/values/{id} [delete]
func (c *deviceCatalogController) DeleteValue() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := c.service.DeleteValue(ctx.Param("id")); err != nil {
			handleDeviceCatalogError(ctx, err)
			return
		}
		ctx.Status(http.StatusNoContent)
	}
}

func handleDeviceCatalogError(ctx *gin.Context, err error) {
	if utils.IsValidationError(err) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	handleErrorResponse(ctx, err)
}
