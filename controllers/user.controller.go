package controllers

import(
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/gin-gonic/gin"
	"github.com/abisaidfarias/lbtechapi/services"
)

// UserController interface
type UserController interface{
	Save(ctx *gin.Context) models.User
	FindAll() []models.User
}

type controller struct{
	services services.UserService
}
// New implement
func New(services services.UserService) UserController {
	return &controller {
		services: services,
	}
}

func (c *controller) Save(ctx *gin.Context) models.User{
	var user models.User
	ctx.BindJSON(&user)
	c.services.Save(user)
	return user
}

func (c *controller) FindAll() []models.User{
	return c.services.FindAll()
}