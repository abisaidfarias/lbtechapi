package main

import (
	"log"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/controllers"
	"github.com/abisaidfarias/lbtechapi/middlewares"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"

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

	testPlanRepository repositories.ITestPlanRepository = repositories.NewTestPlanRepository()
	testPlanService    services.ITestPlanService        = services.NewTestPlanService(testPlanRepository)
	testPlanController controllers.ITestPlanController  = controllers.NewTestPlanController(testPlanService)

	profileRepository repositories.IProfileRepository = repositories.NewProfileRepository()
	profileService    services.IProfileService        = services.NewProfileService(profileRepository, userRepository)
	profileController controllers.IProfileController  = controllers.NewProfileController(profileService)

	companyRepository repositories.ICompanyRepository = repositories.NewCompanyRepository()
	companyService    services.ICompanyService        = services.NewCompanyService(companyRepository)
	companyController controllers.ICompanyController  = controllers.NewCompanyController(companyService)

	brandRepository repositories.IBrandRepository = repositories.NewBrandRepository()
	brandService    services.IBrandService        = services.NewBrandService(brandRepository)
	brandController controllers.IBrandController  = controllers.NewBrandController(brandService)

	countryRepository repositories.ICountryRepository = repositories.NewCountryRepository()
	countryService    services.ICountryService        = services.NewCountryService(countryRepository)
	countryController controllers.ICountryController  = controllers.NewCountryController(countryService)

	deviceRepository repositories.IDeviceRepository = repositories.NewDeviceRepository()
	deviceService    services.IDeviceService        = services.NewDeviceService(deviceRepository)
	deviceController controllers.IDeviceController  = controllers.NewDeviceController(deviceService)

	storageService    services.IStorageService       = services.NewStorageService()
	storageController controllers.IStorageController = controllers.NewStorageController(storageService)
)

func main() {
	server := gin.Default()
	server.Use(middlewares.CORSMiddleware())
	// server.Use(gindump.Dump())

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
			users.GET("", userController.Get())
			users.GET("/profile", userController.GetProfileByID())
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
			testCases.GET(":id", testCaseController.GetById())
			testCases.PUT(":id", testCaseController.Update())
			testCases.PUT(":id/upgrade", testCaseController.Upgrade())
			testCases.DELETE(":id", testCaseController.Delete())
		}
		testPlan := v1.Group("/test-plan")
		{
			testPlan.POST("", testPlanController.Create())
			testPlan.GET("", testPlanController.Get())
			testPlan.GET(":id", testPlanController.GetById())
			testPlan.PUT(":id", testPlanController.Update())
			testPlan.DELETE(":id", testPlanController.Delete())
		}
		profile := v1.Group("/profile")
		{
			profile.POST("", profileController.Create())
			profile.GET("", profileController.Get())
			profile.GET(":id", profileController.GetById())
			profile.PUT(":id", profileController.Update())
			profile.DELETE(":id", profileController.Delete())
		}
		device := v1.Group("/device")
		{
			device.POST("", deviceController.Create())
			device.GET("", deviceController.Get())
			device.GET(":id", deviceController.GetById())
			device.PUT(":id", deviceController.Update())
			device.DELETE(":id", deviceController.Delete())
		}
		company := v1.Group("/company")
		{
			company.POST("", companyController.Create())
			company.GET("", companyController.Get())
		}
		brand := v1.Group("/brand")
		{
			brand.POST("", brandController.Create())
			brand.GET("", brandController.Get())
		}
		country := v1.Group("/country")
		{
			country.POST("", countryController.Create())
			country.GET("", countryController.Get())
		}
		storage := v1.Group("/upload")
		{
			storage.POST("/images", storageController.UploadImage())
		}
	}

	log.Fatal(server.Run(":8080"))
}
