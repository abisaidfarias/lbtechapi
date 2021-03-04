package main

import (
	"log"

	"github.com/abisaidfarias/lbtechapi/controllers"
	"github.com/gin-gonic/gin"
)

// func setupLogOutput() {
// 	f, _ := os.Create("gin.log")
// 	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
// }

var (
	authController controllers.AuthController = controllers.AuthController{}
)

func main() {

	// setupLogOutput()

	server := gin.Default()

	server.POST("/auth/sign-in", authController.SignIn())

	// server.Use(gindump.Dump())
	// server.GET("/user", func(ctx *gin.Context) {
	// 	ctx.JSON(http.StatusOK, userController.FindAll())
	// })
	// server.POST("/user", func(ctx *gin.Context) {
	// 	ctx.JSON(http.StatusOK, userController.Save(ctx))
	// })

	log.Fatal(server.Run(":8080"))
}
