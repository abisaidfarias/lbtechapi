package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/imeivalidator"
	"github.com/gin-gonic/gin"
)

type IImeiValidatorController interface {
	Template() gin.HandlerFunc
	UploadBatch() gin.HandlerFunc
	ListBatches() gin.HandlerFunc
	ListBatchIMEIs() gin.HandlerFunc
	RegisterBatch() gin.HandlerFunc
	DeleteIMEI() gin.HandlerFunc
	DeleteBatch() gin.HandlerFunc
	LookupIMEI() gin.HandlerFunc
	OabiFile() gin.HandlerFunc
}

type imeiValidatorController struct {
	service services.IImeiValidatorService
}

func NewImeiValidatorController(service services.IImeiValidatorService) IImeiValidatorController {
	return &imeiValidatorController{service: service}
}

// Template godoc
// @Summary IMEI Validator: upload template
// @Description Example CSV in OABI format (IMEI;S/N;quantity, no header, CRLF): one Dual SIM and one Single SIM device with fake IMEIs. Internal users only.
// @Tags ImeiValidator
// @Produce text/csv
// @Security Bearer
// @Success 200 {file} file
// @Failure 403 {object} map[string]string
// @Router /imei-validation/validator/template [get]
func (c *imeiValidatorController) Template() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		file, err := c.service.Template(ctx.MustGet("userID").(string))
		if err != nil {
			handleImeiValidatorError(ctx, err)
			return
		}
		sendFile(ctx, imeivalidator.TemplateFileName, imeivalidator.CSVContentType, file)
	}
}

// UploadBatch godoc
// @Summary IMEI Validator: upload a batch
// @Description Parses a .csv (; or ,) or .xlsx with IMEI;S/N;quantity and no header. Validates IMEI (15 digits + Luhn), S/N, quantity and the rows per S/N, and stores only complete devices whose IMEIs are all new, as "ready" in a new batch (batch is null when nothing is new).
// @Tags ImeiValidator
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param file formData file true ".csv or .xlsx file"
// @Success 201 {object} responses.ImeiBatchUploadResult
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /imei-validation/validator/batches [post]
func (c *imeiValidatorController) UploadBatch() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header, err := ctx.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": services.ImeiValidatorErrFileRequired})
			return
		}
		f, err := header.Open()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "the file could not be read"})
			return
		}
		defer f.Close()
		result, err := c.service.UploadBatch(ctx.MustGet("userID").(string), header.Filename, f)
		if err != nil {
			handleImeiValidatorError(ctx, err)
			return
		}
		ctx.JSON(http.StatusCreated, result)
	}
}

// ListBatches godoc
// @Summary IMEI Validator: list batches
// @Description Newest first. search matches the batch code or a batch holding an IMEI starting with those digits.
// @Tags ImeiValidator
// @Produce json
// @Security Bearer
// @Param status query string false "ready | registered"
// @Param search query string false "batch code or IMEI"
// @Param page query int false "page (default 1)"
// @Param limit query int false "page size (default 20, max 200)"
// @Success 200 {object} responses.ImeiBatchPage
// @Failure 400 {object} map[string]string
// @Router /imei-validation/validator/batches [get]
func (c *imeiValidatorController) ListBatches() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		page, limit, ok := pagingFromQuery(ctx)
		if !ok {
			return
		}
		result, err := c.service.ListBatches(ctx.MustGet("userID").(string),
			ctx.Query("status"), ctx.Query("search"), page, limit)
		if err != nil {
			handleImeiValidatorError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// ListBatchIMEIs godoc
// @Summary IMEI Validator: IMEIs of a batch
// @Tags ImeiValidator
// @Produce json
// @Security Bearer
// @Param id path string true "batch id"
// @Param search query string false "IMEI (prefix)"
// @Param page query int false "page (default 1)"
// @Param limit query int false "page size (default 20, max 200)"
// @Success 200 {object} responses.ImeiBatchRecordPage
// @Failure 404 {object} map[string]string
// @Router /imei-validation/validator/batches/{id}/imeis [get]
func (c *imeiValidatorController) ListBatchIMEIs() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		page, limit, ok := pagingFromQuery(ctx)
		if !ok {
			return
		}
		result, err := c.service.ListBatchIMEIs(ctx.MustGet("userID").(string),
			ctx.Param("id"), ctx.Query("search"), page, limit)
		if err != nil {
			handleImeiValidatorError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// RegisterBatch godoc
// @Summary IMEI Validator: mark a batch as registered in OABI
// @Description Marks the batch and all its IMEIs as registered. Idempotent.
// @Tags ImeiValidator
// @Produce json
// @Security Bearer
// @Param id path string true "batch id"
// @Success 200 {object} responses.ImeiBatch
// @Failure 404 {object} map[string]string
// @Router /imei-validation/validator/batches/{id}/register [post]
func (c *imeiValidatorController) RegisterBatch() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := c.service.RegisterBatch(ctx.MustGet("userID").(string), ctx.Param("id"))
		if err != nil {
			handleImeiValidatorError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// DeleteIMEI godoc
// @Summary IMEI Validator: delete an IMEI from a batch
// @Description Only "ready" IMEIs. Deletes the whole device (every IMEI with the same S/N in the batch, listed in deleted_imeis) and the batch when it ends up empty (batch_deleted=true).
// @Tags ImeiValidator
// @Produce json
// @Security Bearer
// @Param id path string true "batch id"
// @Param imei path string true "IMEI"
// @Success 200 {object} responses.ImeiDeleteResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /imei-validation/validator/batches/{id}/imeis/{imei} [delete]
func (c *imeiValidatorController) DeleteIMEI() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := c.service.DeleteIMEI(ctx.MustGet("userID").(string), ctx.Param("id"), ctx.Param("imei"))
		if err != nil {
			handleImeiValidatorError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// DeleteBatch godoc
// @Summary IMEI Validator: delete a batch and its IMEIs
// @Description Only batches with status "ready".
// @Tags ImeiValidator
// @Security Bearer
// @Param id path string true "batch id"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /imei-validation/validator/batches/{id} [delete]
func (c *imeiValidatorController) DeleteBatch() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := c.service.DeleteBatch(ctx.MustGet("userID").(string), ctx.Param("id")); err != nil {
			handleImeiValidatorError(ctx, err)
			return
		}
		ctx.Status(http.StatusNoContent)
	}
}

// LookupIMEI godoc
// @Summary IMEI Validator: look up one IMEI
// @Tags ImeiValidator
// @Produce json
// @Security Bearer
// @Param imei path string true "IMEI"
// @Success 200 {object} responses.ImeiValidatorLookup
// @Router /imei-validation/validator/imeis/{imei} [get]
func (c *imeiValidatorController) LookupIMEI() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := c.service.LookupIMEI(ctx.MustGet("userID").(string), ctx.Param("imei"))
		if err != nil {
			handleImeiValidatorError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// OabiFile godoc
// @Summary IMEI Validator: file ready for OABI
// @Description CSV for OABI: IMEI;S/N;quantity, no header, CRLF, IMEIs of a device together. File name "<batch code>-OABI.csv".
// @Tags ImeiValidator
// @Produce text/csv
// @Security Bearer
// @Param id path string true "batch id"
// @Success 200 {file} file
// @Failure 404 {object} map[string]string
// @Router /imei-validation/validator/batches/{id}/oabi-file [get]
func (c *imeiValidatorController) OabiFile() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		file, name, err := c.service.OabiFile(ctx.MustGet("userID").(string), ctx.Param("id"))
		if err != nil {
			handleImeiValidatorError(ctx, err)
			return
		}
		sendFile(ctx, name, imeivalidator.CSVContentType, file)
	}
}

func sendFile(ctx *gin.Context, name, contentType string, file []byte) {
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition")
	ctx.Data(http.StatusOK, contentType, file)
}

func pagingFromQuery(ctx *gin.Context) (int, int, bool) {
	parse := func(name string) (int, bool) {
		v := ctx.Query(name)
		if v == "" {
			return 0, true
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid %s", name)})
			return 0, false
		}
		return n, true
	}
	page, ok := parse("page")
	if !ok {
		return 0, 0, false
	}
	limit, ok := parse("limit")
	if !ok {
		return 0, 0, false
	}
	return page, limit, true
}

func handleImeiValidatorError(ctx *gin.Context, err error) {
	switch {
	case utils.IsValidationError(err):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, utils.ErrorForbidden):
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrImeiValidatorNotFound):
		msg := err.Error()
		if i := len(msg) - len(": "+services.ErrImeiValidatorNotFound.Error()); i > 0 {
			msg = msg[:i]
		}
		ctx.JSON(http.StatusNotFound, gin.H{"error": msg})
	default:
		handleErrorResponse(ctx, err)
	}
}
