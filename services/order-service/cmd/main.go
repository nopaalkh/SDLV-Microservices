package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"order-service/config"
	"order-service/internal/db"
	"order-service/internal/handler"
	"order-service/internal/middleware"
	"order-service/internal/repository"
	"order-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	cfg.LogConfig()

	// Initialize database
	dbConn, err := db.InitDB(cfg.GetDBConnectionString())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbConn.Close()

	// Initialize repository
	orderRepo := repository.NewPostgreSQLOrderRepository(dbConn.DB)

	// Initialize services
	identityClient := usecase.NewIdentityClient(cfg.IdentityServiceURL, cfg.InternalAPIKey)
	assetService := usecase.NewAssetService(cfg.CatalogServiceURL, cfg.InternalAPIKey)
	emailService := usecase.NewEmailService(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom)
	paymentService := usecase.NewPaymentService(cfg, orderRepo, emailService, identityClient, assetService)

	// Initialize usecase
	orderUsecase := usecase.NewOrderUsecase(orderRepo, assetService, emailService, paymentService, identityClient, cfg.JWTSecret)

	withdrawalRepo := repository.NewPostgreSQLWithdrawalRepository(dbConn.DB)
	withdrawalUsecase := usecase.NewWithdrawalUsecase(withdrawalRepo, orderUsecase)

	// Initialize handler
	orderHandler := handler.NewOrderHandler(orderUsecase)
	adminHandler := handler.NewAdminOrderHandler(orderUsecase)
	withdrawalHandler := handler.NewWithdrawalHandler(withdrawalUsecase)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// API routes
	api := app.Group("/api")
	orders := api.Group("/orders")

	// Midtrans webhook (public)
	orders.Post("/webhook", orderHandler.MidtransWebhook)

	// Internal service routes (protected with API key)
	internalAPIKey := cfg.InternalAPIKey
	orders.Get("/internal/check-purchase", func(c *fiber.Ctx) error {
		key := c.Get("X-Internal-API-Key")
		if key != internalAPIKey {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid or missing internal API key",
			})
		}
		return c.Next()
	}, orderHandler.CheckPurchase)

	// Protected routes (require authentication)
	protected := orders.Group("", middleware.AuthMiddleware(cfg.JWTSecret))
	protected.Post("/", orderHandler.CreateOrder)
	protected.Get("/my-orders", orderHandler.ListOrders)
	protected.Get("/revenue", orderHandler.GetRevenue)
	protected.Get("/:id", orderHandler.GetOrder)
	protected.Put("/:id/pay", orderHandler.ProcessPayment)
	protected.Put("/:id/mock-pay", orderHandler.MockPayment)

	// Withdrawal routes
	withdrawals := api.Group("/withdrawals", middleware.AuthMiddleware(cfg.JWTSecret))
	withdrawals.Post("/", withdrawalHandler.RequestWithdrawal)
	withdrawals.Get("/my-withdrawals", withdrawalHandler.GetMyWithdrawals)

	// Admin routes (require ADMIN role)
	admin := api.Group("/admin/orders", middleware.AuthMiddleware(cfg.JWTSecret), middleware.RoleMiddleware("ADMIN"))
	admin.Get("", adminHandler.ListAllOrders)
	admin.Get("/revenue", adminHandler.GetPlatformRevenue)

	adminWithdrawals := api.Group("/admin/withdrawals", middleware.AuthMiddleware(cfg.JWTSecret), middleware.RoleMiddleware("ADMIN"))
	adminWithdrawals.Get("", withdrawalHandler.GetAllWithdrawals)
	adminWithdrawals.Put("/:id/approve", withdrawalHandler.ApproveWithdrawal)
	adminWithdrawals.Put("/:id/reject", withdrawalHandler.RejectWithdrawal)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Start server in a goroutine
	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}

	log.Println("Server exited")
}
