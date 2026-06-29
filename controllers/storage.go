package controllers

import (
	"io"
	"mime/multipart"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IStorageController interface {
	UploadFile() gin.HandlerFunc
}

// AuthController implementation of the interface
type storageController struct {
	storageService services.IStorageService
}
type Form struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

//NewUserController is the constructor
func NewStorageController(storageService services.IStorageService) IStorageController {
	return &storageController{
		storageService: storageService,
	}
}

// UploadFile godoc
// @Summary Cargar archivo
// @Description Sube un archivo al storage (máximo 50 MB). La key en S3 es única; la respuesta incluye el nombre original del archivo.
// @Tags Storage
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param file formData file true "Archivo a subir (máx. 50 MB)"
// @Success 201 {object} responses.UploadFileResponse "Archivo subido"
// @Failure 400 {object} map[string]string "Archivo inválido"
// @Failure 401 {object} map[string]string "No autorizado"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /upload/images [post]
func (c *storageController) UploadFile() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var form Form
		err := ctx.ShouldBind(&form)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := utils.ValidateUploadFileSize(form.File.Size); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fileContent, err := form.File.Open()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer fileContent.Close()

		byteContainer, err := io.ReadAll(io.LimitReader(fileContent, utils.MaxUploadFileSize+1))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if int64(len(byteContainer)) > utils.MaxUploadFileSize {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "file exceeds maximum size of 50MB"})
			return
		}

		result, err := c.storageService.UploadUserFile(byteContainer, form.File.Filename)
		if err != nil {
			if utils.IsValidationError(err) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return
		}
		ctx.JSON(http.StatusCreated, result)
	}
}
