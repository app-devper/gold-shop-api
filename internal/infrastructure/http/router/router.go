package router

import (
	"github.com/devper-gold/gold-shop-api/internal/infrastructure/http/handler"
	"github.com/devper-gold/gold-shop-api/internal/infrastructure/http/middleware"
	"github.com/devper-gold/gold-shop-api/pkg/jwt"
	"github.com/gin-gonic/gin"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	Auth       *handler.AuthHandler
	Branch     *handler.BranchHandler
	Sale       *handler.SaleHandler
	Pawn       *handler.PawnHandler
	GoldPrice  *handler.GoldPriceHandler
	GoldSaving *handler.GoldSavingHandler
	User       *handler.UserHandler
	Product    *handler.ProductHandler
	Customer   *handler.CustomerHandler
	Inventory  *handler.InventoryHandler
	Reward     *handler.RewardHandler
	Expense    *handler.ExpenseHandler
	Report     *handler.ReportHandler
}

// Setup sets up all routes
func Setup(r *gin.Engine, jwtManager *jwt.Manager, handlers *Handlers) {
	// Middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	api := r.Group("/api/v1")

	// Auth routes (public)
	auth := api.Group("/auth")
	{
		auth.POST("/login", handlers.Auth.Login)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(jwtManager))
	{
		// Auth
		protected.POST("/auth/logout", handlers.Auth.Logout)
		protected.POST("/auth/refresh", handlers.Auth.RefreshToken)
		protected.GET("/auth/me", handlers.Auth.GetCurrentUser)
		protected.PUT("/auth/password", handlers.Auth.ChangePassword)

		// Branches
		branches := protected.Group("/branches")
		{
			branches.GET("", handlers.Branch.List)
			branches.GET("/:id", handlers.Branch.Get)
			branches.POST("", middleware.RoleMiddleware("admin"), handlers.Branch.Create)
			branches.PUT("/:id", middleware.RoleMiddleware("admin"), handlers.Branch.Update)
			branches.DELETE("/:id", middleware.RoleMiddleware("admin"), handlers.Branch.Delete)
		}

		// Gold Prices
		goldPrices := protected.Group("/gold-prices")
		{
			goldPrices.GET("/current", handlers.GoldPrice.GetCurrent)
			goldPrices.GET("/history", handlers.GoldPrice.GetHistory)
			goldPrices.POST("", middleware.RoleMiddleware("admin", "manager"), handlers.GoldPrice.SetPrice)
			goldPrices.POST("/sync", middleware.RoleMiddleware("admin", "manager"), handlers.GoldPrice.SyncFromAPI)
		}

		// Sales
		sales := protected.Group("/sales")
		{
			sales.GET("", handlers.Sale.List)
			sales.GET("/unpaid", handlers.Sale.GetUnpaid)
			sales.GET("/:id", handlers.Sale.Get)
			sales.POST("", handlers.Sale.Create)
			sales.POST("/:id/cancel", handlers.Sale.Cancel)
			sales.GET("/:id/receipt", handlers.Sale.GetReceipt)
		}

		// Pawns
		pawns := protected.Group("/pawns")
		{
			pawns.GET("", handlers.Pawn.List)
			pawns.GET("/due-soon", handlers.Pawn.GetDueSoon)
			pawns.GET("/:id", handlers.Pawn.Get)
			pawns.POST("", handlers.Pawn.Create)
			pawns.POST("/:id/pay-interest", handlers.Pawn.PayInterest)
			pawns.POST("/:id/redeem", handlers.Pawn.Redeem)
			pawns.POST("/:id/extend", handlers.Pawn.Extend)
			pawns.POST("/:id/forfeit", handlers.Pawn.Forfeit)
		}

		// Gold Savings
		goldSavings := protected.Group("/gold-savings")
		{
			goldSavings.GET("", handlers.GoldSaving.List)
			goldSavings.GET("/:id", handlers.GoldSaving.Get)
			goldSavings.POST("", handlers.GoldSaving.OpenAccount)
			goldSavings.POST("/:id/deposit", handlers.GoldSaving.Deposit)
			goldSavings.POST("/:id/withdraw", handlers.GoldSaving.Withdraw)
			goldSavings.POST("/:id/close", handlers.GoldSaving.Close)
			goldSavings.GET("/:id/statement", handlers.GoldSaving.GetStatement)
		}

		// Users
		users := protected.Group("/users")
		users.Use(middleware.RoleMiddleware("admin"))
		{
			users.POST("", handlers.User.CreateUser)
			users.GET("", handlers.User.GetUsers)
			users.GET("/:id", handlers.User.GetUser)
			users.PUT("/:id", handlers.User.UpdateUser)
			users.DELETE("/:id", handlers.User.DeleteUser)
			users.POST("/:id/reset-password", handlers.User.ResetPassword)
		}

		// Product Categories
		productCategories := protected.Group("/product-categories")
		{
			productCategories.GET("", handlers.Product.GetCategories)
			productCategories.POST("", middleware.RoleMiddleware("admin", "manager"), handlers.Product.CreateCategory)
		}

		// Products
		products := protected.Group("/products")
		{
			products.GET("", handlers.Product.GetProducts)
			products.GET("/:id", handlers.Product.GetProduct)
			products.POST("", middleware.RoleMiddleware("admin", "manager"), handlers.Product.CreateProduct)
			products.PUT("/:id", middleware.RoleMiddleware("admin", "manager"), handlers.Product.UpdateProduct)
			products.DELETE("/:id", middleware.RoleMiddleware("admin", "manager"), handlers.Product.DeleteProduct)
		}

		// Customers
		customers := protected.Group("/customers")
		{
			customers.GET("", handlers.Customer.GetCustomers)
			customers.GET("/:id", handlers.Customer.GetCustomer)
			customers.GET("/rfid/:rfid", handlers.Customer.GetByRFID)
			customers.POST("", handlers.Customer.CreateCustomer)
			customers.PUT("/:id", handlers.Customer.UpdateCustomer)
			customers.DELETE("/:id", middleware.RoleMiddleware("admin", "manager"), handlers.Customer.DeleteCustomer)
		}

		// Inventory Transfers
		inventory := protected.Group("/inventory")
		{
			inventory.GET("/transfers", handlers.Inventory.GetTransfers)
			inventory.GET("/transfers/:id", handlers.Inventory.GetTransfer)
			inventory.POST("/transfers", handlers.Inventory.CreateTransfer)
			inventory.POST("/transfers/:id/approve", handlers.Inventory.ApproveTransfer)
			inventory.POST("/transfers/:id/receive", handlers.Inventory.ReceiveTransfer)
			inventory.POST("/transfers/:id/cancel", handlers.Inventory.CancelTransfer)
		}

		// Rewards
		rewards := protected.Group("/rewards")
		{
			rewards.GET("", handlers.Reward.GetRewards)
			rewards.POST("", middleware.RoleMiddleware("admin", "manager"), handlers.Reward.CreateReward)
			rewards.POST("/redeem", handlers.Reward.RedeemReward)
		}

		// Expenses
		expenses := protected.Group("/expenses")
		{
			expenses.GET("", handlers.Expense.GetExpenses)
			expenses.POST("", handlers.Expense.CreateExpense)
			expenses.GET("/categories", handlers.Expense.GetCategories)
			expenses.POST("/categories", middleware.RoleMiddleware("admin"), handlers.Expense.CreateCategory)
		}

		// Reports
		reports := protected.Group("/reports")
		{
			reports.GET("/dashboard", handlers.Report.GetDashboardData)
			reports.GET("/profit-loss", handlers.Report.GetProfitLossReport)
			reports.GET("/multi-branch", middleware.RoleMiddleware("admin"), handlers.Report.GetMultiBranchProfitLossReport)
			reports.GET("/top-products", handlers.Report.GetTopProducts)
			reports.GET("/employee-performance", handlers.Report.GetEmployeePerformance)
			reports.GET("/trends", handlers.Report.GetSalesTrends)
		}
	}
}
