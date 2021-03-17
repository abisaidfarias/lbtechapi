package main

import (
	"log"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/controllers"
	"github.com/abisaidfarias/lbtechapi/middlewares"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/gin-contrib/cors"

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

	testCategoryRepository repositories.ITestCategoryRepository = repositories.NewTestCategoryRepository()
	testCategoryService    services.ITestCategoryService        = services.NewTestCategoryService(testCategoryRepository)
	testCategoryController controllers.ITestCategoryController  = controllers.NewTestCategoryController(testCategoryService)

	testCaseRepository repositories.ITestCaseRepository = repositories.NewTestCaseRepository()
	testCaseService    services.ITestCaseService        = services.NewTestCaseService(testCaseRepository)
	testCaseController controllers.ITestCaseController  = controllers.NewTestCaseController(testCaseService)
)

func main() {
	server := gin.Default()
	// server.Use(gindump.Dump())

	server.Use(cors.Default())

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("passwordFormat", utils.ValidPassword)
	} else {
		panic("error validation")
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("testCaseCode", utils.ValidTestCaseCode)
	} else {
		panic("error validation")
	}

	v1 := server.Group("/api/v1")
	{
		v1.GET("/health", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"status": "server is up"})
		})

		v1.POST("/sign-in", authController.SignIn())
		v1.Use(middlewares.AuthMiddleware())

		users := v1.Group("/users")
		{
			users.POST("", userController.Create())
			users.GET(":id", userController.GetByID())
			users.PUT(":id", userController.Update())
			users.DELETE(":id", userController.Delete())
		}

		categories := v1.Group("/test-categories")
		{
			categories.POST("", testCategoryController.Create())
			categories.GET("", testCategoryController.Get())
		}

		testCases := v1.Group("/test-cases")
		{
			testCases.POST("", testCaseController.Create())
			testCases.GET("", testCaseController.Get())
			testCases.GET(":id", testCaseController.GetByID())
			testCases.PUT(":id", testCaseController.Update())
			testCases.PUT(":id/upgrade", testCaseController.Upgrade())
			testCases.DELETE(":id", testCaseController.Delete())
		}
	}

	log.Fatal(server.Run(":8080"))
}
