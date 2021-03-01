package main

import (
	"github.com/abisaidfarias/lbtechapi/controllers"
	"github.com/abisaidfarias/lbtechapi/services"
	"net/http"
	"github.com/gin-gonic/gin"
)

var (
	userService services.UserService = services.New()
	userController controllers.UserController = controllers.New(userService)
)

func main(){
	server := gin.Default()
	server.GET("/user",func(ctx *gin.Context){
		ctx.JSON(http.StatusOK,userController.FindAll())
	})
	server.POST("/user",func(ctx *gin.Context){
		ctx.JSON(http.StatusOK,userController.Save(ctx))
	})
	server.Run(":8089")
}