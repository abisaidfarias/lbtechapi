package main

import (
	"log"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/controllers"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"

	// "github.com/abisaidfarias/lbtechapi/middlewares"
	"github.com/abisaidfarias/lbtechapi/models"
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

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("passwordFormat", utils.ValidPassword)
	}

	auth := server.Group("/api/v1")
	{
		auth.POST("/sign-in", authController.SignIn())
		auth.POST("/create", userController.Create())
	}

	// server.Use(middlewares.AuthMiddleware())
	// server.Use(testAuthMiddleware())

	server.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "server is up"})
	})

	log.Fatal(server.Run(":8080"))
}

func testAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		requestUser, exists := c.Get("user")

		if !exists {
			log.Printf("Unable to extract user from request context for unknown reason: %v\n", c)
		} else {
			user := requestUser.(*models.User)
			log.Println(user.ID)
			log.Println(user.Email)
		}
		// before request
		c.Next()

	}
}
