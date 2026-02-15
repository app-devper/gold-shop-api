package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devper-gold/gold-shop-api/config"
	"github.com/devper-gold/gold-shop-api/internal/application/auth"
	"github.com/devper-gold/gold-shop-api/internal/application/branch"
	"github.com/devper-gold/gold-shop-api/internal/application/customer"
	"github.com/devper-gold/gold-shop-api/internal/application/expense"
	gold_price_app "github.com/devper-gold/gold-shop-api/internal/application/gold_price"
	"github.com/devper-gold/gold-shop-api/internal/application/gold_saving"
	"github.com/devper-gold/gold-shop-api/internal/application/inventory"
	"github.com/devper-gold/gold-shop-api/internal/application/pawn"
	"github.com/devper-gold/gold-shop-api/internal/application/product"
	"github.com/devper-gold/gold-shop-api/internal/application/report"
	"github.com/devper-gold/gold-shop-api/internal/application/reward"
	"github.com/devper-gold/gold-shop-api/internal/application/sale"
	"github.com/devper-gold/gold-shop-api/internal/application/user"
	"github.com/devper-gold/gold-shop-api/internal/domain/entity"
	"github.com/devper-gold/gold-shop-api/internal/domain/repository"
	gold_price_infra "github.com/devper-gold/gold-shop-api/internal/infrastructure/external/gold_price"
	"github.com/devper-gold/gold-shop-api/internal/infrastructure/http/handler"
	"github.com/devper-gold/gold-shop-api/internal/infrastructure/http/router"
	"github.com/devper-gold/gold-shop-api/internal/infrastructure/mongo"
	"github.com/devper-gold/gold-shop-api/pkg/jwt"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to MongoDB
	mongoClient, err := mongo.NewClient(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := mongoClient.Close(ctx); err != nil {
			log.Printf("Error closing MongoDB connection: %v", err)
		}
	}()

	log.Println("Connected to MongoDB")

	// Initialize repositories
	branchRepo := mongo.NewBranchRepository(mongoClient)
	userRepo := mongo.NewUserRepository(mongoClient)
	customerRepo := mongo.NewCustomerRepository(mongoClient)
	productRepo := mongo.NewProductRepository(mongoClient)
	saleRepo := mongo.NewSaleRepository(mongoClient)
	pawnRepo := mongo.NewPawnRepository(mongoClient)
	goldPriceRepo := mongo.NewGoldPriceRepository(mongoClient)
	goldSavingRepo := mongo.NewGoldSavingRepository(mongoClient)

	productCategoryRepo := mongo.NewProductCategoryRepository(mongoClient)
	productItemRepo := mongo.NewProductItemRepository(mongoClient)
	stockLogRepo := mongo.NewStockLogRepository(mongoClient)
	expenseCategoryRepo := mongo.NewExpenseCategoryRepository(mongoClient)
	expenseRepo := mongo.NewExpenseRepository(mongoClient)
	inventoryRepo := mongo.NewInventoryTransferRepository(mongoClient)
	rewardRepo := mongo.NewRewardRepository(mongoClient)
	redemptionRepo := mongo.NewRewardRedemptionRepository(mongoClient)

	// Initialize external API clients (before default data init)
	goldAPIClient := gold_price_infra.NewThaiGoldAPIClient(cfg.GoldAPI.URL)

	// Initialize default data (branch, admin user, and gold price from API)
	ctx := context.Background()
	if err := initializeDefaultData(ctx, branchRepo, userRepo, goldPriceRepo, goldAPIClient); err != nil {
		log.Printf("Warning: Failed to initialize default data: %v", err)
	}

	// Initialize JWT manager
	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.ExpirationHours)

	// Initialize services
	authService := auth.NewService(userRepo, jwtManager)
	branchService := branch.NewService(branchRepo)
	saleService := sale.NewService(saleRepo, productRepo, productItemRepo, goldPriceRepo, stockLogRepo, customerRepo, branchRepo, userRepo)
	pawnService := pawn.NewService(pawnRepo, branchRepo)
	goldPriceService := gold_price_app.NewService(goldPriceRepo, goldAPIClient)
	goldSavingService := gold_saving.NewService(goldSavingRepo, goldPriceRepo, branchRepo)
	// Suppress unused variables until their services are implemented

	userService := user.NewService(userRepo, branchRepo)
	productService := product.NewService(productRepo, productItemRepo, stockLogRepo, productCategoryRepo, branchRepo)
	customerService := customer.NewService(customerRepo)
	inventoryService := inventory.NewService(inventoryRepo, productRepo, branchRepo)
	rewardService := reward.NewService(rewardRepo, redemptionRepo, customerRepo)
	expenseService := expense.NewService(expenseRepo, expenseCategoryRepo)
	reportService := report.NewService(saleRepo, pawnRepo, expenseRepo, goldSavingRepo, branchRepo)

	// Initialize handlers
	handlers := &router.Handlers{
		Auth:       handler.NewAuthHandler(authService),
		Branch:     handler.NewBranchHandler(branchService),
		Sale:       handler.NewSaleHandler(saleService),
		Pawn:       handler.NewPawnHandler(pawnService),
		GoldPrice:  handler.NewGoldPriceHandler(goldPriceService),
		GoldSaving: handler.NewGoldSavingHandler(goldSavingService),
		User:       handler.NewUserHandler(userService),
		Product:    handler.NewProductHandler(productService),
		Customer:   handler.NewCustomerHandler(customerService),
		Inventory:  handler.NewInventoryHandler(inventoryService),
		Reward:     handler.NewRewardHandler(rewardService),
		Expense:    handler.NewExpenseHandler(expenseService),
		Report:     handler.NewReportHandler(reportService),
	}

	// Setup router
	r := gin.Default()
	router.Setup(r, jwtManager, handlers)

	// Start server
	go func() {
		addr := ":" + cfg.Server.Port
		log.Printf("Server starting on %s", addr)
		if err := r.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}

// initializeDefaultData seeds the database with a default branch, admin user, and gold price
func initializeDefaultData(ctx context.Context, branchRepo repository.BranchRepository, userRepo repository.UserRepository, goldPriceRepo repository.GoldPriceRepository, goldAPIClient gold_price_app.ExternalGoldPriceAPI) error {
	// Check if default branch exists
	defaultBranch, err := branchRepo.GetByCode(ctx, "HQ")
	if err == entity.ErrNotFound {
		// Create default branch
		defaultBranch = entity.NewBranch("HQ", "สำนักงานใหญ่", "กรุงเทพมหานคร", "02-000-0000")
		if err := branchRepo.Create(ctx, defaultBranch); err != nil {
			return err
		}
		log.Println("Created default branch: HQ - สำนักงานใหญ่")
	} else if err != nil {
		return err
	}

	// Check if admin user exists
	_, err = userRepo.GetByUsername(ctx, "admin")
	if err == entity.ErrNotFound {
		// Hash the default password
		passwordHash, err := utils.HashPassword("admin123")
		if err != nil {
			return err
		}

		// Create admin user
		adminUser := entity.NewUser(
			defaultBranch.ID,
			"EMP001",
			"admin",
			passwordHash,
			"System Administrator",
			entity.RoleAdmin,
		)
		if err := userRepo.Create(ctx, adminUser); err != nil {
			return err
		}
		log.Println("Created default admin user: admin (please change password immediately)")
	} else if err != nil {
		return err
	}

	// Check if gold price exists
	_, err = goldPriceRepo.GetCurrent(ctx)
	if err == entity.ErrNotFound {
		// Fetch gold price from external API
		priceData, err := goldAPIClient.GetCurrentPrice(ctx)
		if err != nil {
			log.Printf("Warning: Failed to fetch gold price from API: %v, using default values", err)
			// Fallback to default values if API fails
			priceData = &gold_price_app.GoldPriceData{
				GoldBarBuy:       42350.00,
				GoldBarSell:      42450.00,
				GoldOrnamentBuy:  41850.00,
				GoldOrnamentSell: 42950.00,
			}
		}

		// Create gold price from API data
		goldPrice := entity.NewGoldPrice(
			priceData.GoldBarBuy,
			priceData.GoldBarSell,
			priceData.GoldOrnamentBuy,
			priceData.GoldOrnamentSell,
			"api",
		)
		if err := goldPriceRepo.Create(ctx, goldPrice); err != nil {
			return err
		}
		log.Printf("Created gold price from API: Bar(Buy:%.2f/Sell:%.2f) Ornament(Buy:%.2f/Sell:%.2f)",
			priceData.GoldBarBuy, priceData.GoldBarSell, priceData.GoldOrnamentBuy, priceData.GoldOrnamentSell)
	} else if err != nil {
		return err
	}

	return nil
}
