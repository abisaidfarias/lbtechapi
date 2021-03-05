package main

import (
	"log"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/controllers"
	"github.com/abisaidfarias/lbtechapi/middlewares"
	"github.com/gin-gonic/gin"
)

var (
	authController controllers.AuthController = controllers.AuthController{}
)

func main() {

	server := gin.Default()

	server.Use(middlewares.AuthMiddleware())

	auth := server.Group("/auth")
	{
		auth.POST("/sign-in", authController.SignIn())
		auth.POST("/sign-up", authController.SignUp())
	}

	server.GET("/status", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "server is up"})
	})

	log.Fatal(server.Run(":8080"))
}
