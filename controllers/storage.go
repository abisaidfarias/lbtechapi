package controllers

import (
	"log"
	"mime/multipart"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IStorageController interface {
	UploadImage() gin.HandlerFunc
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
func (c *storageController) UploadImage() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		log.Println("llego")
		var form Form

		_ = ctx.ShouldBind(&form)
		log.Println("llego 2")

		ctx.String(http.StatusOK, "Files uploaded")
		return
	}
}
