package controllers

import (
	"net/http"

	"github.com/abisaidfarias/lbtechapi/services"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/gin-gonic/gin"
)

// IUserController controller
type IPersonController interface {
	Create() gin.HandlerFunc
	Get() gin.HandlerFunc
	Delete() gin.HandlerFunc
	Update() gin.HandlerFunc
}

// AuthController implementation of the interface
type personController struct {
	personService services.IPersonService
}

//NewUserController is the constructor
func NewPersonController(personService services.IPersonService) IPersonController {
	return &personController{
		personService: personService,
	}
}

// Create creates a category
func (c *personController) Create() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var person request.Person

		err := ctx.ShouldBindJSON(&person)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := c.personService.Create(&person)

		if err != nil {
			if utils.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utils.ErrorDuplicated.Error()})
				return
			}
			ctx.Status(http.StatusInternalServerError)
			handleErrorResponse(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"id": id})

	}
}

// Get list all categories
func (c *personController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		companies, err := c.personService.Get()

		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, companies)
	}

}
func (c *personController) Delete() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var id string = ctx.Param("id")

		hasRelations, err := c.personService.Delete(id)
		if err != nil {
			handleErrorResponse(ctx, err)
			return
		}
		if hasRelations {
			ctx.Status(http.StatusConflict)
			return
		}
		ctx.Status(http.StatusOK)
	}
}
func (c *personController) Update() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var id string = ctx.Param("id")

		var person request.Person

		err := ctx.ShouldBindJSON(&person)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = c.personService.Update(id, &person)

		if err != nil {
			if utils.ErrorDuplicatedData(err) {
				ctx.JSON(http.StatusConflict, gin.H{"error": utils.ErrorDuplicated.Error()})
				return
			}
			handleErrorResponse(ctx, err)
			return

		}

		ctx.Status(http.StatusOK)
	}
}
