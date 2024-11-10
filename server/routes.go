package server

import (
	"daijai/controllers"
	"daijai/middlewares"
	"daijai/models"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://daijai-fe-f4almcs4iq-as.a.run.app"},
		AllowMethods:     []string{http.MethodPost, http.MethodOptions, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodGet},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "*"},
		AllowCredentials: true,
	}))

	//Routes for healthcheck of api server
	healthcheck := router.Group("health")
	{
		health := new(controllers.HealthController)
		healthcheck.GET("/", health.Status)
		healthcheck.GET("/ping", health.Ping)
	}

	//Routes for swagger
	// swagger := router.Group("swagger")
	// {
	// 	docs.SwaggerInfo.Title = "Golang REST API Starter"
	// 	docs.SwaggerInfo.Description = "This is a sample backend written in Go."
	// 	docs.SwaggerInfo.Version = "1.0"
	// 	docs.SwaggerInfo.Host = "cloudfactory.swagger.io"
	// 	docs.SwaggerInfo.BasePath = "/v1"
	// 	swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// }

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Routes"})
		// c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	categories := router.Group("categories")
	{
		categoryController := controllers.NewCategoryController(db)
		categories.POST("", categoryController.CreateCategory)
		categories.GET("", categoryController.GetCategories)
		categories.GET("/:id", categoryController.GetCategoryByID)
		categories.PUT("/:id", categoryController.UpdateCategory)
		categories.DELETE("/:id", categoryController.DeleteCategory)
	}

	projectStore := router.Group("projectStores")
	{
		projectStoreController := controllers.NewProjectStoreController(db)
		projectStore.POST("", projectStoreController.CreateProjectStore)
		projectStore.GET("", projectStoreController.GetProjectStores)
		projectStore.GET("/:id", projectStoreController.GetProjectStoreByID)
		projectStore.PUT("/:id", projectStoreController.UpdateProjectStore)
		projectStore.DELETE("/:id", projectStoreController.DeleteProjectStore)
	}

	inventories := router.Group("inventories")
	{
		inventoryController := controllers.NewInventoryController(db)
		inventories.POST("", inventoryController.CreateInventory)
		inventories.GET("", inventoryController.GetInventories)
		inventories.GET("/:slug", inventoryController.GetInventoryBySlug)
		inventories.PUT("/:id", inventoryController.UpdateInventory)
		inventories.DELETE("/:id", inventoryController.DeleteInventory)
		inventories.POST("/transfer", inventoryController.TransferMaterial)
		inventories.POST("/transfer/calculateCost", inventoryController.CalculateCostOfTransferMaterial)
		// inventories.GET("/:id/transactions/:material_id", inventoryController.GetInventoryTransaction)
		// inventories.GET("/transactions", inventoryController.GetAllInventoryTransactions)
	}

	/// transacton routes
	transactions := router.Group("transactions")
	{
		transactionController := controllers.NewTransactionController(db)
		transactions.GET("", transactionController.GetTransactions)
		transactions.GET("/inventories", transactionController.GetTransactionsGroupByInventory)
		transactions.GET("/po/:poNumber", transactionController.GetInventoryMaterialTransactionsByPONumber)
		// transactions.POST("", transactionController.CreateTransaction)
		// transactions.GET("/:id", transactionController.GetTransactionByID)
		// transactions.PUT("/:id", transactionController.UpdateTransaction)
		// transactions.DELETE("/:id", transactionController.DeleteTransaction)
	}

	materials := router.Group("materials")
	{
		materialController := controllers.NewMaterialController(db)
		materials.POST("", materialController.CreateMaterial)
		materials.GET("/trash", materialController.GetDeletedMaterials)
		materials.PUT("/restore/:id", materialController.RestoreMaterial)
		materials.DELETE("/destroy/:id", materialController.PermanentlyDeleteMaterial)
		materials.GET("", materialController.GetMaterials)
		materials.GET("/query", materialController.QueryMaterials)
		materials.GET("/:slug", materialController.GetMaterialBySlug)
		materials.PUT("/:id", materialController.UpdateMaterial)
		materials.PUT("/adjust/:id", materialController.AdjustMaterialQuantity)
		materials.DELETE("/:id", materialController.DeleteMaterial)
		materials.GET("/search", materialController.SearchMaterials)
	}

	drawings := router.Group("drawings")
	{
		drawingCtrl := controllers.NewDrawingController(db)
		drawings.GET("/new/info/:type", drawingCtrl.GetNewDrawingInfo)
		drawings.POST("", drawingCtrl.CreateDrawing)
		drawings.GET("", drawingCtrl.GetDrawings)
		drawings.GET("/:id", drawingCtrl.GetDrawingByID)
		drawings.PUT("/:id", drawingCtrl.UpdateDrawing)
		drawings.DELETE("/:id", drawingCtrl.DeleteDrawing)
	}

	orders := router.Group("orders")
	{
		ctrl := controllers.NewOrderController(db)
		orders.POST("", ctrl.CreateOrder)
		orders.GET("", ctrl.GetOrders)
		orders.GET("/:slug", ctrl.GetOrderBySlug)
		orders.GET("/info/:slug", ctrl.GetOrderInfo)
		orders.GET("/new/info", ctrl.GetNewOrderInfo)

		extCtrl := controllers.NewExtendOrdererController(db)
		ext := orders.Group("extenders")
		{
			ext.GET("", extCtrl.GetExtendOrders)
			ext.GET("/new/info", extCtrl.GetNewInfo)
			ext.GET("/:slug", extCtrl.GetExtendOrderBySlug)
			ext.POST("", extCtrl.CreateExtendOrders)
		}
	}

	withdrawals := router.Group("withdrawals")
	{
		withdrawCtrl := controllers.NewWithdrawalController(db)
		withdrawals.POST("/orderReserving/adjust/:id", withdrawCtrl.AdjustOrderReserving)
		withdrawals.GET("/new/info", withdrawCtrl.GetNewWithdrawInfo)
		withdrawals.POST("admin", withdrawCtrl.CreateWithdrawalAdmin)
		withdrawals.POST("", withdrawCtrl.CreateWithdrawal)
		withdrawals.POST("/admin/nonspec/withdraw",
			middlewares.AuthMiddleware(models.ROLE_Admin, models.ROLE_Manager),
			withdrawCtrl.CreateNonSpecificOrderWithdrawal)
		withdrawals.POST("/partial", withdrawCtrl.CreatePartialWithdrawal)
		withdrawals.GET("", withdrawCtrl.GetAllWithdrawals)
		withdrawals.PUT("/:id", withdrawCtrl.UpdateWithdrawal)
		withdrawals.GET("/:slug", withdrawCtrl.GetWithdrawalBySlug)
		withdrawals.DELETE("/:id", withdrawCtrl.DeleteWithdraw)
		withdrawals.GET("/new/admin/info", middlewares.AuthMiddleware(models.ROLE_Admin), withdrawCtrl.GetNewWithdrawAdminInfo)
		withdrawals.PUT("/approve/:id",
			middlewares.AuthMiddleware(models.ROLE_Admin, models.ROLE_Manager),
			withdrawCtrl.ApproveWithdrawal)
	}

	pr := router.Group("pr")
	{
		ctrl := controllers.NewPurchaseRequisitionController(db)
		pr.POST("", ctrl.CreatePurchaseRequisition)
		pr.GET("", ctrl.GetAllPurchaseRequisition)
		pr.GET("/new/info", ctrl.GetNewPRInfo)
		pr.GET("/:slug", ctrl.GetPurchaseRequisition)
		pr.PUT("/:id", ctrl.UpdatePurchaseRequisition)
		pr.PUT("/approve/:slug", ctrl.ApprovePurchaseRequisition)
		pr.DELETE("/:id", ctrl.DeletePurchaseRequisition)
	}

	projects := router.Group("projects")
	{
		// projects.Use(middlewares.AuthMiddleware("technician", "admin", "user"))
		ctrl := controllers.NewProjectController(db)
		projects.POST("", ctrl.CreateProject)
		projects.GET("", ctrl.GetAllProjects)
		projects.GET("/:id", ctrl.GetProject)
		projects.GET("/detail/:slug", ctrl.GetProjectDetailBySlug)
		projects.PUT("/:id", ctrl.UpdateProject)
		projects.DELETE("/:id", ctrl.DeleteProject)
	}

	receipts := router.Group("receipts")
	{
		ctrl := controllers.NewReceipt(db)
		receipts.POST("", ctrl.CreateReceipt)
		receipts.GET("", ctrl.GetAllReceipts)
		receipts.GET("/new/info", ctrl.GetNewReceiptInfo)
		receipts.GET("/edit/info/:slug", ctrl.GetEditReceiptInfo)
		receipts.GET("/details/:slug", ctrl.GetReceiptBySlug)
		receipts.PUT("/:slug", ctrl.UpdateReceipt)
		receipts.DELETE("/:id", ctrl.DeleteReceipt)
		receipts.PUT("/approve/:id", ctrl.ApproveReceipt)
	}

	users := router.Group("users")
	{
		userCtrl := controllers.NewUser(db)
		users.POST("", userCtrl.CreateUser)
		users.PUT("/reset/:id", userCtrl.ResetPassword)
		users.GET("", userCtrl.GetAllUsers)
		users.GET("/:id", userCtrl.GetUser)
		users.PUT("/:id", userCtrl.UpdateUser)
		users.DELETE("/:id", userCtrl.DeleteUser)
	}

	auth := router.Group("auth")
	{
		authCtrl := controllers.NewAuth(db)
		auth.POST("/register", authCtrl.Register)
		auth.POST("/login", authCtrl.Login)
		auth.POST("/logout", authCtrl.Logout)
		auth.GET("/session", authCtrl.Session)
	}

	slugs := router.Group("slugs")
	{
		ctrl := controllers.NewSlugController(db)
		slugs.GET("", ctrl.GetAllSluggers)
		slugs.GET("/request/:slug", ctrl.RequestSlug)
		slugs.GET("/:slug", ctrl.GetSlug)
	}

	notifications := router.Group("notifications")
	{
		ctrl := controllers.NewNotificationController(db)
		notifications.GET("", ctrl.GetNotifications)
	}

	imageCtrl := controllers.NewImageController()
	images := router.Group("images")
	{
		images.POST("/upload/:directory", imageCtrl.Upload)
		images.GET("/:directory/:fileName", imageCtrl.Download)
	}

	plannerCtrl := controllers.NewPlannerController(db)
	planner := router.Group("planner")
	{
		planner.GET("/new/info", plannerCtrl.GetNewPlannerInfo)
		planner.GET("/extend/orders", plannerCtrl.GetExtendOrders)
		planner.GET("/materials", plannerCtrl.GetMaterialSumByInventory)
		planner.POST("/confirm", plannerCtrl.CreatePlanner)
		planner.POST("/inquiry", plannerCtrl.InquiryPlan)
	}

	filters := router.Group("filters")
	{
		filterController := controllers.NewFilterController(db)
		filters.GET("/categories", filterController.GetCategories)
		filters.GET("/categories/:id/materials", filterController.GetMaterialsByCategoryID)
	}

	return router
}
