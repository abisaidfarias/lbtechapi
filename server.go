package main

import (
	"log"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/controllers"
	"github.com/abisaidfarias/lbtechapi/middlewares"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/gin-gonic/gin"
)

var (
	authController controllers.AuthController = controllers.AuthController{}
)

func main() {

	server := gin.Default()

	// server.Use(gindump.Dump())

	auth := server.Group("/auth")
	{
		auth.POST("/sign-in", authController.SignIn())
		auth.POST("/sign-up", authController.SignUp())
	}

	server.Use(middlewares.AuthMiddleware())
	server.Use(testAuthMiddleware())

	server.GET("/status", func(ctx *gin.Context) {
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
