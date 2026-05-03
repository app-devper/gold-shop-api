package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devper-gold/gold-shop-api/app/config"
	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/feature/branch"
	"github.com/devper-gold/gold-shop-api/app/feature/customer"
	"github.com/devper-gold/gold-shop-api/app/feature/employee"
	"github.com/devper-gold/gold-shop-api/app/feature/expense"
	gold_price_app "github.com/devper-gold/gold-shop-api/app/feature/gold_price"
	"github.com/devper-gold/gold-shop-api/app/feature/gold_saving"
	"github.com/devper-gold/gold-shop-api/app/feature/inventory"
	"github.com/devper-gold/gold-shop-api/app/feature/pawn"
	"github.com/devper-gold/gold-shop-api/app/feature/product"
	"github.com/devper-gold/gold-shop-api/app/feature/report"
	"github.com/devper-gold/gold-shop-api/app/feature/reward"
	"github.com/devper-gold/gold-shop-api/app/feature/sale"
	gold_price_infra "github.com/devper-gold/gold-shop-api/app/infrastructure/external/gold_price"
	"github.com/devper-gold/gold-shop-api/app/infrastructure/http/handler"
	"github.com/devper-gold/gold-shop-api/app/infrastructure/http/router"
	"github.com/devper-gold/gold-shop-api/app/infrastructure/mongo"
	redisrepo "github.com/devper-gold/gold-shop-api/app/infrastructure/redis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type App struct{}

func (a App) StartApp() {
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to MongoDB
	mongoClient, err := mongo.NewClient(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		logrus.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := mongoClient.Close(ctx); err != nil {
			logrus.Errorf("Error closing MongoDB connection: %v", err)
		}
	}()
	logrus.Info("Connected to MongoDB")

	// Connect to Redis (um-api shared session store)
	redisOpt, err := redis.ParseURL(cfg.Redis.Host)
	if err != nil {
		logrus.Fatalf("Failed to parse Redis URL: %v", err)
	}
	rdb := redis.NewClient(redisOpt)
	defer rdb.Close()
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logrus.Fatalf("Failed to connect to Redis: %v", err)
	}
	logrus.Info("Connected to Redis")

	// Initialize repositories
	branchRepo := mongo.NewBranchRepository(mongoClient)
	employeeRepo := mongo.NewEmployeeRepository(mongoClient)
	sessionRepo := redisrepo.NewSessionRepository(rdb)
	customerRepo := mongo.NewCustomerRepository(mongoClient)
	productRepo := mongo.NewProductRepository(mongoClient)
	saleRepo := mongo.NewSaleRepository(mongoClient)
	pawnRepo := mongo.NewPawnRepository(mongoClient)
	goldPriceRepo := mongo.NewGoldPriceRepository(mongoClient)
	goldSavingRepo := mongo.NewGoldSavingRepository(mongoClient)
	productItemRepo := mongo.NewProductItemRepository(mongoClient)
	stockLogRepo := mongo.NewStockLogRepository(mongoClient)
	expenseCategoryRepo := mongo.NewExpenseCategoryRepository(mongoClient)
	expenseRepo := mongo.NewExpenseRepository(mongoClient)
	inventoryRepo := mongo.NewInventoryTransferRepository(mongoClient)
	rewardRepo := mongo.NewRewardRepository(mongoClient)
	redemptionRepo := mongo.NewRewardRedemptionRepository(mongoClient)

	// Initialize external API clients
	goldAPIClient := gold_price_infra.NewThaiGoldAPIClient(cfg.GoldAPI.URL)

	// Seed default data
	ctx := context.Background()
	if err := initializeDefaultData(ctx, branchRepo, goldPriceRepo, goldAPIClient); err != nil {
		logrus.Warnf("Failed to initialize default data: %v", err)
	}

	// Initialize services
	employeeService := employee.NewService(employeeRepo, branchRepo)
	branchService := branch.NewService(branchRepo)
	txManager := mongo.NewTransactionManager(mongoClient)
	saleService := sale.NewService(saleRepo, productRepo, productItemRepo, goldPriceRepo, stockLogRepo, customerRepo, branchRepo, nil, txManager)
	pawnService := pawn.NewService(pawnRepo, branchRepo)
	goldPriceService := gold_price_app.NewService(goldPriceRepo, goldAPIClient)
	goldSavingService := gold_saving.NewService(goldSavingRepo, goldPriceRepo, branchRepo, customerRepo)
	productService := product.NewService(productRepo, productItemRepo, stockLogRepo, branchRepo)
	customerService := customer.NewService(customerRepo)
	inventoryService := inventory.NewService(inventoryRepo, productRepo, branchRepo, txManager)
	rewardService := reward.NewService(rewardRepo, redemptionRepo, customerRepo, txManager)
	expenseService := expense.NewService(expenseRepo, expenseCategoryRepo)
	reportService := report.NewService(saleRepo, pawnRepo, expenseRepo, goldSavingRepo, branchRepo)

	// Initialize handlers
	handlers := &router.Handlers{
		Branch:     handler.NewBranchHandler(branchService),
		Employee:   handler.NewEmployeeHandler(employeeService),
		Sale:       handler.NewSaleHandler(saleService),
		Pawn:       handler.NewPawnHandler(pawnService),
		GoldPrice:  handler.NewGoldPriceHandler(goldPriceService),
		GoldSaving: handler.NewGoldSavingHandler(goldSavingService),
		Product:    handler.NewProductHandler(productService),
		Customer:   handler.NewCustomerHandler(customerService),
		Inventory:  handler.NewInventoryHandler(inventoryService),
		Reward:     handler.NewRewardHandler(rewardService),
		Expense:    handler.NewExpenseHandler(expenseService),
		Report:     handler.NewReportHandler(reportService),
	}

	// Setup router
	r := gin.New()
	if err := r.SetTrustedProxies(nil); err != nil {
		logrus.Error(err)
	}
	router.Setup(r, cfg.Auth.SecretKey, sessionRepo, employeeRepo, branchRepo, handlers)

	// Start server
	srv := &http.Server{Addr: ":" + cfg.Server.Port, Handler: r}
	go func() {
		logrus.Infof("Server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logrus.Errorf("Server forced shutdown: %v", err)
	}
	logrus.Info("Server exited")
}

func initializeDefaultData(ctx context.Context, branchRepo interface {
	GetByCode(ctx context.Context, code string) (*entity.Branch, error)
	Create(ctx context.Context, branch *entity.Branch) error
}, goldPriceRepo interface {
	GetCurrent(ctx context.Context) (*entity.GoldPrice, error)
	Create(ctx context.Context, price *entity.GoldPrice) error
}, goldAPIClient gold_price_app.ExternalGoldPriceAPI) error {
	_, err := branchRepo.GetByCode(ctx, "HQ")
	if err == entity.ErrNotFound {
		defaultBranch := entity.NewBranch("HQ", "สำนักงานใหญ่", "กรุงเทพมหานคร", "02-000-0000")
		if err := branchRepo.Create(ctx, defaultBranch); err != nil {
			return err
		}
		logrus.Info("Created default branch: HQ - สำนักงานใหญ่")
	} else if err != nil {
		return err
	}

	_, err = goldPriceRepo.GetCurrent(ctx)
	if err == entity.ErrNotFound {
		priceData, err := goldAPIClient.GetCurrentPrice(ctx)
		if err != nil {
			logrus.Warnf("Failed to fetch gold price from API: %v, using default values", err)
			priceData = &gold_price_app.GoldPriceData{
				GoldBarBuy:       42350.00,
				GoldBarSell:      42450.00,
				GoldOrnamentBuy:  41850.00,
				GoldOrnamentSell: 42950.00,
			}
		}
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
		logrus.Infof("Created gold price from API: Bar(Buy:%.2f/Sell:%.2f) Ornament(Buy:%.2f/Sell:%.2f)",
			priceData.GoldBarBuy, priceData.GoldBarSell, priceData.GoldOrnamentBuy, priceData.GoldOrnamentSell)
	} else if err != nil {
		return err
	}

	return nil
}
