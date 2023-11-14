package server

import (
	"daijai/controllers"
	"daijai/docs"
	"net/http"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"   // gin-swagger middleware
	"github.com/swaggo/gin-swagger/swaggerFiles" // swagger embed files
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	//Routes for healthcheck of api server
	healthcheck := router.Group("health")
	{
		health := new(controllers.HealthController)
		healthcheck.GET("/health", health.Status)
		healthcheck.GET("/ping", health.Ping)
	}

	//Routes for swagger
	swagger := router.Group("swagger")
	{
		// programatically set swagger info
		docs.SwaggerInfo.Title = "Golang REST API Starter"
		docs.SwaggerInfo.Description = "This is a sample backend written in Go."
		docs.SwaggerInfo.Version = "1.0"
		// docs.SwaggerInfo.Host = "cloudfactory.swagger.io"
		// docs.SwaggerInfo.BasePath = "/v1"

		swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	router.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	materials := router.Group("materials")
	{
		materialController := controllers.NewMaterialController(db)
		materials.POST("/", materialController.CreateMaterial)
		materials.GET("/", materialController.GetMaterials)
		materials.GET("/:id", materialController.GetMaterialByID)
		materials.PUT("/:id", materialController.UpdateMaterial)
		materials.DELETE("/:id", materialController.DeleteMaterial)
	}

	drawings := router.Group("drawings")
	{
		drawingCtrl := controllers.NewDrawingController(db)
		drawings.POST("/", drawingCtrl.CreateDrawing)
		drawings.GET("/", drawingCtrl.GetDrawings)
		drawings.GET("/:id", drawingCtrl.GetDrawingByID)
		// drawings.PUT("/:id", drawingCtrl.UpdateDrawing)
		drawings.DELETE("/:id", drawingCtrl.DeleteDrawing)
	}

	withdrawals := router.Group("withdrawals")
	{
		withdrawCtrl := controllers.NewWithdrawalController(db)
		withdrawals.POST("/", withdrawCtrl.CreateWithdrawal)
		withdrawals.GET("/", withdrawCtrl.GetAllWithdrawals)
		withdrawals.PUT("/approve/:id", withdrawCtrl.ApproveWithdrawal)
		// withdrawals.GET("/:id", drawingCtrl.GetDrawingByID)
		// withdrawals.DELETE("/:id", drawingCtrl.DeleteDrawing)
	}

	pr := router.Group("pr")
	{
		ctrl := controllers.NewPurchaseRequisitionController(db)
		pr.POST("/", ctrl.CreatePurchaseRequisition)
		pr.GET("/", ctrl.GetAllPurchaseRequisition)
		pr.GET("/:id", ctrl.GetAllPurchaseRequisition)
		pr.PUT("/:id", ctrl.UpdatePurchaseRequisition)
		pr.DELETE("/:id", ctrl.DeletePurchaseRequisition)
	}

	projects := router.Group("projects")
	{
		ctrl := controllers.NewProjectController(db)
		projects.POST("/", ctrl.CreateProject)
		projects.GET("/", ctrl.GetAllProjects)
		projects.GET("/:id", ctrl.GetProject)
		projects.PUT("/:id", ctrl.UpdateProject)
		projects.DELETE("/:id", ctrl.DeleteProject)
	}

	receipts := router.Group("receipts")
	{
		ctrl := controllers.NewReceipt(db)
		receipts.POST("/", ctrl.CreateReceipt)
		receipts.GET("/", ctrl.GetAllReceipts)
		receipts.GET("/:id", ctrl.GetReceipt)
		receipts.PUT("/:id", ctrl.UpdateReceipt)
		receipts.DELETE("/:id", ctrl.DeleteReceipt)
	}

	return router

}
