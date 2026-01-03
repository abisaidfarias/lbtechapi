package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/abisaidfarias/lbtechapi/config"
	"github.com/abisaidfarias/lbtechapi/controllers"
	"github.com/abisaidfarias/lbtechapi/middlewares"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/services"
	"github.com/abisaidfarias/lbtechapi/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/abisaidfarias/lbtechapi/docs"
)

// init loads secrets before any global variables are initialized
func init() {
	log.Println("🔐 Loading secrets...")
	_, err := config.LoadSecrets()
	if err != nil {
		log.Fatal("Failed to load secrets:", err)
	}
	log.Println("✅ Secrets loaded successfully")
}

var (
	userRepository repositories.IUserRepository = repositories.NewUserRepository()
	authService    services.IAuthService        = services.NewAuthService(userRepository)
	authController controllers.IAuthController  = controllers.NewAuthController(authService)

	profileRepository repositories.IProfileRepository = repositories.NewProfileRepository()
	companyRepository repositories.ICompanyRepository = repositories.NewCompanyRepository()
	
	userService    services.IUserService       = services.NewUserService(userRepository, profileRepository, companyRepository)
	userController controllers.IUserController = controllers.NewUserController(userService)

	testCategoryRepository repositories.ITestCategoryRepository = repositories.NewTestCategoryRepository()
	testCategoryService    services.ITestCategoryService        = services.NewTestCategoryService(testCategoryRepository)
	testCategoryController controllers.ITestCategoryController  = controllers.NewTestCategoryController(testCategoryService)

	testCaseRepository repositories.ITestCaseRepository = repositories.NewTestCaseRepository()
	testCaseService    services.ITestCaseService        = services.NewTestCaseService(testCaseRepository, testCategoryService)
	testCaseController controllers.ITestCaseController  = controllers.NewTestCaseController(testCaseService)

	profileService    services.IProfileService        = services.NewProfileService(profileRepository, userRepository)
	profileController controllers.IProfileController  = controllers.NewProfileController(profileService)
	companyService    services.ICompanyService        = services.NewCompanyService(companyRepository, homologationRepository,
		deviceTrackingRepository, userRepository)
	companyController controllers.ICompanyController = controllers.NewCompanyController(companyService)

	brandRepository repositories.IBrandRepository = repositories.NewBrandRepository()
	brandService    services.IBrandService        = services.NewBrandService(brandRepository)
	brandController controllers.IBrandController  = controllers.NewBrandController(brandService)

	countryRepository repositories.ICountryRepository = repositories.NewCountryRepository()
	countryService    services.ICountryService        = services.NewCountryService(countryRepository, homologationRepository)
	countryController controllers.ICountryController  = controllers.NewCountryController(countryService)

	deviceRepository repositories.IDeviceRepository = repositories.NewDeviceRepository()
	deviceService    services.IDeviceService        = services.NewDeviceService(deviceRepository,
		deviceTrackingRepository, homologationRepository, brandRepository, userRepository)
	deviceController controllers.IDeviceController = controllers.NewDeviceController(deviceService)

	storageService    services.IStorageService       = services.NewStorageService()
	storageController controllers.IStorageController = controllers.NewStorageController(storageService)

	homologationRepository repositories.IHomologationRepository = repositories.NewHomologationRepository()
	homologationService    services.IHomologationService        = services.NewHomologationService(homologationRepository, testCategoryRepository,
		userRepository, notificationRepository, brandRepository, deviceRepository, countryRepository)
	homologationController controllers.IHomologationController = controllers.NewHomologationController(homologationService)

	dashboardService    services.IDashboardService       = services.NewDashboardService(homologationRepository, userRepository)
	dashboardController controllers.IDashboardController = controllers.NewDashboardController(dashboardService)

	kpiService    services.IKpiService       = services.NewKpiService(homologationRepository, userRepository)
	kpiController controllers.IKpiController = controllers.NewKpiController(kpiService)

	printerRepository repositories.IPrinterRepository = repositories.NewPrinterRepository()
	printerService    services.IPrinterService        = services.NewPrinterService(printerRepository)
	printerController controllers.IPrinterController  = controllers.NewPrinterController(printerService)

	locationRepository repositories.ILocationRepository = repositories.NewLocationRepository()
	locationService    services.ILocationService        = services.NewLocationService(locationRepository)
	locationController controllers.ILocationController  = controllers.NewLocationController(locationService)

	deviceTrackingRepository repositories.IDeviceTrackingRepository = repositories.NewDeviceTrackingRepository()
	deviceTrackingService    services.IDeviceTrackingService        = services.NewDeviceTrackingService(deviceTrackingRepository,
		userRepository, companyRepository, brandRepository, deviceRepository, countryRepository)
	deviceTrackingController controllers.IDeviceTrackingController = controllers.NewDeviceTrackingController(deviceTrackingService)

	testPlanRepository repositories.ITestPlanRepository = repositories.NewTestPlanRepository()
	testPlanService    services.ITestPlanService        = services.NewTestPlanService(testPlanRepository, homologationRepository)
	testPlanController controllers.ITestPlanController  = controllers.NewTestPlanController(testPlanService)

	configurationRepository repositories.IConfigurationRepository = repositories.NewConfigurationRepository()
	configurationService    services.IConfigurationService        = services.NewConfigurationService(configurationRepository)
	configurationController controllers.IConfigurationController  = controllers.NewConfigurationController(configurationService)

	personRepository repositories.IPersonRepository = repositories.NewPersonRepository()
	personService    services.IPersonService        = services.NewPersonService(personRepository, deviceTrackingRepository)
	personController controllers.IPersonController  = controllers.NewPersonController(personService)

	notificationRepository repositories.INotificationRepository = repositories.NewNotificationRepository()
	notificationService    services.INotificationService        = services.NewNotificationService(notificationRepository)
	notificationController controllers.INotificationController  = controllers.NewNotificationController(notificationService)
)

// @title LBTech API
// @version 1.0
// @description API para gestión de homologaciones y dispositivos
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@lbtech.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	server := gin.Default()
	server.Use(middlewares.CORSMiddleware())

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
	server.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": fmt.Sprintf("server is up %s", config.GetValue("MONGO_DB"))})
	})
	
	// Swagger endpoint
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	v1 := server.Group("/api/v1")
	{
		v1.GET("/health", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"status": "server is up"})
		})
		printer := v1.Group("/printer")
		{
			printer.GET("", printerController.Get())
			printer.POST("", printerController.Create())
		}
		upgrade := v1.Group("/upgrade")
		{
			upgrade.POST("", userController.Upgrade())
		}
		v1.POST("/sign-in", authController.SignIn())
		v1.Use(middlewares.AuthMiddleware())

		users := v1.Group("/users")
		{
			users.POST("", userController.Create())
			users.GET("", userController.Get())
			users.GET("/profile", userController.GetProfileByID())
			users.PUT(":id", userController.Update())
			users.DELETE(":id", userController.Delete())
			users.PUT(":id/changePassword", userController.ChangePassword())
			users.GET("/company/:id", userController.GetUserByCompany())
			users.GET("/internal", userController.GetByInternal())
		}

		categories := v1.Group("/test-categories")
		{
			categories.POST("", testCategoryController.Create())
			categories.GET("", testCategoryController.Get())
		}

		testCases := v1.Group("/test-cases")
		{
			testCases.POST("", testCaseController.Create())
			testCases.POST("fileUpload", testCaseController.FileUpload())
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
			company.PUT(":id", companyController.Update())
			company.DELETE(":id", companyController.Delete())
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
			country.PUT(":id", countryController.Update())
			country.DELETE(":id", countryController.Delete())
		}
		storage := v1.Group("/upload")
		{
			storage.POST("/images", storageController.UploadFile())
		}
		homologation := v1.Group("/homologation")
		{
			homologation.POST("", homologationController.Create())
			homologation.GET("", homologationController.Get())
			homologation.GET(":id/report", homologationController.GetReport())
			homologation.GET(":id/categories/test", homologationController.GetCategoriesWithTest())
			homologation.PUT(":id", homologationController.UpdateTestResult())
			homologation.PUT(":id/phase", homologationController.PhaseChange())
			homologation.GET(":id/test/fails", homologationController.GetHomologationFails())
			homologation.POST(":id/failTest", homologationController.CreateFailTest())
			homologation.PUT(":id/document", homologationController.UpdateDocument())
			homologation.PUT(":id/homologation", homologationController.Update())
			homologation.DELETE(":id", homologationController.Delete())
			homologation.PUT(":id/failTest", homologationController.UpdateFailTest())

		}
		export := v1.Group("/export")
		{
			export.GET("/homologation", homologationController.ExportHomologation())
			export.POST("/device-tracking", deviceTrackingController.ExportDeviceTracking())
			export.GET("/fail-test/:id", homologationController.ExportFailTest())
		}
		dashboard := v1.Group("/dashboard")
		{
			dashboard.GET("", dashboardController.Get())
			dashboard.GET("/info", dashboardController.GetGeneralInfo())
		}
		kpi := v1.Group("/kpi")
		{
			kpi.GET("/volume/:start/:end", kpiController.GetVolumeChart())
			kpi.GET("/time/:start/:end", kpiController.GetTimeChart())
		}
		location := v1.Group("/location")
		{
			location.POST("", locationController.Create())
			location.GET("", locationController.Get())
		}
		deviceTracking := v1.Group("/device-tracking")
		{
			deviceTracking.POST("", deviceTrackingController.Create())
			deviceTracking.GET("", deviceTrackingController.Get())
			deviceTracking.PUT("", deviceTrackingController.AddTrakingLog())
			deviceTracking.DELETE(":id", deviceTrackingController.Delete())
			deviceTracking.PUT(":id", deviceTrackingController.Update())
			deviceTracking.POST("advanced-search", deviceTrackingController.AdvancedSearch())
			deviceTracking.GET("search-options", deviceTrackingController.AdvancedSearchOptions())
		}
		configuration := v1.Group("/configuration")
		{
			configuration.POST("", configurationController.Create())
			configuration.GET("", configurationController.Get())
		}
		person := v1.Group("/person")
		{
			person.POST("", personController.Create())
			person.GET("", personController.Get())
			person.PUT(":id", personController.Update())
			person.DELETE(":id", personController.Delete())
		}
		notification := v1.Group("/notification")
		{
			notification.POST("", notificationController.Create())
			notification.GET("/company/:id", notificationController.GetByCompany())
		}
	}

	log.Fatal(server.Run(":8080"))
}
