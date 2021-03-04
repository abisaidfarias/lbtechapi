package controllers

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/gin-gonic/gin"
)

// UserController interface
type UserController interface {
	Save(ctx *gin.Context) error
	FindAll() []models.User
}

type controller struct {
	services services.UserService
}

// New implement
func New(services services.UserService) UserController {
	return &controller{
		services: services,
	}
}

func (c *controller) Save(ctx *gin.Context) error {
	var user models.User
	err := ctx.BindJSON(&user)
	if err != nil {
		return err
	}
	c.services.Save(user)
	return nil

}

func (c *controller) FindAll() []models.User {
	return c.services.FindAll()
}
