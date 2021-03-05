package main

import (
	"log"

	"github.com/abisaidfarias/lbtechapi/controllers"
	"github.com/gin-gonic/gin"
)

var (
	authController controllers.AuthController = controllers.AuthController{}
)

func main() {

	server := gin.Default()

	auth := server.Group("/auth")
	{
		auth.POST("/sign-in", authController.SignIn())
		auth.POST("/sign-up", authController.SignUp())
	}

	log.Fatal(server.Run(":8080"))
}
