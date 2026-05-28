package router

import (
	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"github.com/devper-gold/gold-shop-api/app/infrastructure/http/handler"
	"github.com/devper-gold/gold-shop-api/app/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	Branch     *handler.BranchHandler
	Sale       *handler.SaleHandler
	Pawn       *handler.PawnHandler
	GoldPrice  *handler.GoldPriceHandler
	GoldSaving *handler.GoldSavingHandler
	Employee   *handler.EmployeeHandler
	Product    *handler.ProductHandler
	Customer   *handler.CustomerHandler
	Inventory  *handler.InventoryHandler
	Reward     *handler.RewardHandler
	Expense    *handler.ExpenseHandler
	Report     *handler.ReportHandler
}

// Setup sets up all routes
func Setup(r *gin.Engine, secretKey string, sessionRepo middleware.SessionLookup, employeeRepo repository.EmployeeRepository, branchRepo repository.BranchRepository, handlers *Handlers) {

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

	// Protected routes: RequireAuthenticated → RequireTenant → RequireSession → RequireBranch
	protected := api.Group("")
	protected.Use(
		middleware.RequireAuthenticated(secretKey),
		middleware.RequireTenant(),
		middleware.RequireSession(sessionRepo),
		middleware.RequireBranch(employeeRepo, branchRepo),
	)
	{
		// Branches
		branches := protected.Group("/branches")
		{
			branches.GET("", handlers.Branch.List)
			branches.GET("/:id", handlers.Branch.Get)
			branches.POST("", middleware.RoleMiddleware(entity.EmployeeRoleAdmin), handlers.Branch.Create)
			branches.PUT("/:id", middleware.RoleMiddleware(entity.EmployeeRoleAdmin), handlers.Branch.Update)
			branches.DELETE("/:id", middleware.RoleMiddleware(entity.EmployeeRoleAdmin), handlers.Branch.Delete)
		}

		// Employees
		employees := protected.Group("/employees")
		{
			employees.GET("", handlers.Employee.GetEmployees)
			employees.GET("/me", handlers.Employee.GetMyEmployee)
			employees.GET("/branch/:branchId", handlers.Employee.GetEmployeesByBranch)
			employees.GET("/:id", handlers.Employee.GetEmployee)
			employees.POST("", middleware.RequireRole(entity.UMRoleSuper, entity.UMRoleAdmin), handlers.Employee.CreateEmployee)
			employees.PUT("/:id", middleware.RequireRole(entity.UMRoleSuper, entity.UMRoleAdmin), handlers.Employee.UpdateEmployee)
			employees.DELETE("/:id", middleware.RequireRole(entity.UMRoleSuper, entity.UMRoleAdmin), handlers.Employee.DeleteEmployee)
		}

		// Gold Prices
		goldPrices := protected.Group("/gold-prices")
		{
			goldPrices.GET("/current", handlers.GoldPrice.GetCurrent)
			goldPrices.GET("/history", handlers.GoldPrice.GetHistory)
			goldPrices.POST("", middleware.RoleMiddleware(entity.EmployeeRoleAdmin, entity.EmployeeRoleManager), handlers.GoldPrice.SetPrice)
			goldPrices.POST("/sync", middleware.RoleMiddleware(entity.EmployeeRoleAdmin, entity.EmployeeRoleManager), handlers.GoldPrice.SyncFromAPI)
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
			// Admin-only manual correction (signed weight delta).
			goldSavings.POST("/:id/adjust", middleware.RoleMiddleware(entity.EmployeeRoleAdmin), handlers.GoldSaving.Adjust)
		}

		// Products (catalog masters + per-piece items)
		products := protected.Group("/products")
		{
			mgr := middleware.RoleMiddleware(entity.EmployeeRoleAdmin, entity.EmployeeRoleManager)
			products.GET("", handlers.Product.GetProducts)
			products.GET("/:id", handlers.Product.GetProduct)
			products.POST("", mgr, handlers.Product.CreateProduct)
			products.PUT("/:id", mgr, handlers.Product.UpdateProduct)
			products.DELETE("/:id", mgr, handlers.Product.DeleteProduct)

			// Item-level (each gold piece has its own barcode + status)
			products.GET("/:id/items", handlers.Product.ListItems)
			products.POST("/:id/items", mgr, handlers.Product.CreateItem)
			products.POST("/:id/items/bulk", mgr, handlers.Product.BulkCreateItems)
			products.PUT("/:id/items/:itemId", mgr, handlers.Product.UpdateItem)
			products.DELETE("/:id/items/:itemId", mgr, handlers.Product.DeleteItem)
		}

		// Customers
		customers := protected.Group("/customers")
		{
			customers.GET("", handlers.Customer.GetCustomers)
			customers.GET("/:id", handlers.Customer.GetCustomer)
			customers.GET("/rfid/:rfid", handlers.Customer.GetByRFID)
			customers.POST("", handlers.Customer.CreateCustomer)
			customers.PUT("/:id", handlers.Customer.UpdateCustomer)
			customers.DELETE("/:id", middleware.RoleMiddleware(entity.EmployeeRoleAdmin, entity.EmployeeRoleManager), handlers.Customer.DeleteCustomer)
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
			rewards.POST("", middleware.RoleMiddleware(entity.EmployeeRoleAdmin, entity.EmployeeRoleManager), handlers.Reward.CreateReward)
			rewards.POST("/redeem", handlers.Reward.RedeemReward)
		}

		// Expenses
		expenses := protected.Group("/expenses")
		{
			expenses.GET("", handlers.Expense.GetExpenses)
			expenses.POST("", handlers.Expense.CreateExpense)
			expenses.GET("/categories", handlers.Expense.GetCategories)
			expenses.POST("/categories", middleware.RoleMiddleware(entity.EmployeeRoleAdmin), handlers.Expense.CreateCategory)
		}

		// Reports
		reports := protected.Group("/reports")
		{
			reports.GET("/dashboard", handlers.Report.GetDashboardData)
			reports.GET("/profit-loss", handlers.Report.GetProfitLossReport)
			reports.GET("/multi-branch", middleware.RoleMiddleware(entity.EmployeeRoleAdmin), handlers.Report.GetMultiBranchProfitLossReport)
			reports.GET("/top-products", handlers.Report.GetTopProducts)
			reports.GET("/employee-performance", handlers.Report.GetEmployeePerformance)
			reports.GET("/trends", handlers.Report.GetSalesTrends)
		}
	}
}
