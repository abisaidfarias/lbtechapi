package controllers

import (
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
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

// Create creates a category
func (c *storageController) UploadFile() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var form Form
		err := ctx.ShouldBind(&form)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fileContent, _ := form.File.Open()
		byteContainer, err := ioutil.ReadAll(fileContent)
		if err != nil {
			return
		}
		url, err := c.storageService.UploadFile(byteContainer)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}
		ctx.JSON(http.StatusCreated, gin.H{"url": url})
	}
}
