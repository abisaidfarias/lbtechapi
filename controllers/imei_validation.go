package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

type IImeiValidationController interface {
	Lookup() gin.HandlerFunc
	ListHistory() gin.HandlerFunc
	ExportHistory() gin.HandlerFunc
}

type imeiValidationController struct {
	service services.IImeiValidationService
}

func NewImeiValidationController(service services.IImeiValidationService) IImeiValidationController {
	return &imeiValidationController{service: service}
}

func (c *imeiValidationController) Lookup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)
		var body request.ImeiLookup
		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := c.service.Lookup(userID, body.IMEI)
		if err != nil {
			handleImeiValidationError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func (c *imeiValidationController) ListHistory() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)
		filter, err := imeiHistoryFilterFromQuery(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		page, err := c.service.ListHistory(userID, filter)
		if err != nil {
			handleImeiValidationError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, page)
	}
}

func (c *imeiValidationController) ExportHistory() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.MustGet("userID").(string)
		filter, err := imeiHistoryFilterFromQuery(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		file, err := c.service.ExportHistory(userID, filter)
		if err != nil {
			handleImeiValidationError(ctx, err)
			return
		}
		ctx.Header("Content-Disposition", "attachment; filename=imei-queries.xlsx")
		ctx.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", file)
	}
}

func imeiHistoryFilterFromQuery(ctx *gin.Context) (repositories.ImeiValidationListFilter, error) {
	filter := repositories.ImeiValidationListFilter{
		IMEI:           ctx.Query("imei"),
		Brand:          ctx.Query("brand"),
		TechnicalModel: ctx.Query("technical_model"),
		CommercialName: ctx.Query("commercial_name"),
		Source:         ctx.Query("source"),
	}
	if v := ctx.Query("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return filter, err
		}
		filter.Page = n
	}
	if v := ctx.Query("page_size"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return filter, err
		}
		filter.PageSize = n
	}
	if v := ctx.Query("date_from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return filter, err
		}
		filter.DateFrom = &t
	}
	if v := ctx.Query("date_to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return filter, err
		}
		filter.DateTo = &t
	}
	return filter, nil
}

func handleImeiValidationError(ctx *gin.Context, err error) {
	switch {
	case utils.IsValidationError(err):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, utils.ErrorForbidden):
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, utils.ErrorImeiInfoRateLimited):
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
	default:
		handleErrorResponse(ctx, err)
	}
}
