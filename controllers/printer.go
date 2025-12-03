package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IPrinterController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
}

// AuthController implementation of the interface
type printerController struct {
	printerService services.IPrinterService
}

//NewUserController is the constructor
func NewPrinterController(printerService services.IPrinterService) IPrinterController {
	return &printerController{
		printerService: printerService,
	}
}

// Create godoc
// @Summary Crear nueva impresora
// @Description Crea una nueva impresora en el sistema
// @Tags Printers
// @Accept json
// @Produce json
// @Param printer body request.Printer true "Datos de la impresora"
// @Success 201 "Impresora creada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 409 {object} map[string]string "Impresora duplicada"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /printer [post]
func (c *printerController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var printer request.Printer

		err := ctx.ShouldBindJSON(&printer)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.printerService.Create(&printer)

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
	}
}

// Get godoc
// @Summary Listar impresoras
// @Description Obtiene la lista de todas las impresoras
// @Tags Printers
// @Accept json
// @Produce json
// @Success 200 {array} responses.Printer "Lista de impresoras"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /printer [get]
func (c *printerController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		printers, err := c.printerService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, printers)
	}

}
