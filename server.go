package main

import (
	"io"
	"net/http"
	"os"

	"github.com/abisaidfarias/lbtechapi/controllers"
	"github.com/abisaidfarias/lbtechapi/middlewares"
	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/gin-gonic/gin"
	gindump "github.com/tpkeeper/gin-dump"
)

func setupLogOutput() {
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
}

var (
	userService    services.UserService       = services.New()
	userController controllers.UserController = controllers.New(userService)
)

func main() {
	setupLogOutput()

	server := gin.New()

	server.Use(gin.Recovery(), middlewares.Logger(),
		middlewares.BasicAuth(), gindump.Dump())

	server.GET("/user", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, userController.FindAll())
	})
	server.POST("/user", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, userController.Save(ctx))

	})
	server.Run(":8089")
}
