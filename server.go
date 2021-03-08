package main

import (
	"log"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/controllers"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"

	// "github.com/abisaidfarias/lbtechapi/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var (
	userRepository repositories.IUserRepository = repositories.NewUserRepository()
	authService    services.IAuthService        = services.NewAuthService(userRepository)
	authController controllers.IAuthController  = controllers.NewAuthController(authService)

	userService    services.IUserService       = services.NewUserService(userRepository)
	userController controllers.IUserController = controllers.NewUserController(userService)
)

func main() {

	server := gin.Default()

	// server.Use(gindump.Dump())
	// server.Use(middlewares.AuthMiddleware())

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("passwordFormat", utils.ValidPassword)
	}

	auth := server.Group("/api/v1")
	{
		auth.POST("/sign-in", authController.SignIn())
		auth.POST("/create", userController.Create())
	}

	server.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "server is up"})
	})

	log.Fatal(server.Run(":8080"))
}
