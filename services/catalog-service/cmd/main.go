package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"catalog-service/config"
	"catalog-service/internal/db"
	"catalog-service/internal/handler"
	"catalog-service/internal/middleware"
	"catalog-service/internal/repository"
	"catalog-service/internal/usecase"

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
	assetRepo := repository.NewPostgreSQLAssetRepository(dbConn.DB)
	reviewRepo := repository.NewPostgreSQLReviewRepository(dbConn.DB)

	// Initialize usecase
	orderClient := usecase.NewOrderClient(cfg.OrderServiceURL, cfg.InternalAPIKey)
	identityClient := usecase.NewIdentityClient(cfg.IdentityServiceURL, cfg.InternalAPIKey)
	assetUsecase := usecase.NewAssetUsecase(assetRepo)
	reviewUsecase := usecase.NewReviewUsecase(reviewRepo, orderClient, identityClient)

	// Initialize handler
	assetHandler := handler.NewAssetHandler(assetUsecase, orderClient)
	adminHandler := handler.NewAdminAssetHandler(assetUsecase)
	reviewHandler := handler.NewReviewHandler(reviewUsecase)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024, // 50 MB
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
	assets := api.Group("/assets")

	// Public routes (no auth required)
	assets.Post("/search", assetHandler.SearchAssets)

	// Protected route registered first (before /:id wildcard)
	assets.Get("/my-assets", middleware.AuthMiddleware(cfg.JWTSecret), assetHandler.ListAssets)

	// Internal service routes (protected with API key)
	internalAPIKey := cfg.InternalAPIKey
	assets.Put("/internal/:id/increment-sales", func(c *fiber.Ctx) error {
		key := c.Get("X-Internal-API-Key")
		if key != internalAPIKey {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid or missing internal API key",
			})
		}
		return c.Next()
	}, assetHandler.IncrementSalesCount)

	// Public wildcard route (must be after specific named routes)
	assets.Get("/:id", assetHandler.GetAsset)
	assets.Get("/:assetId/reviews", reviewHandler.GetReviews)

	// Protected routes (require auth)
	assets.Post("/:assetId/reviews", middleware.AuthMiddleware(cfg.JWTSecret), reviewHandler.CreateReview)

	// Auth middleware for remaining routes
	assets.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	// Protected routes
	assets.Post("/", assetHandler.CreateAsset)
	assets.Put("/:id", assetHandler.UpdateAsset)
	assets.Delete("/:id", assetHandler.DeleteAsset)
	assets.Get("/:id/download-link", assetHandler.GetDownloadLink)

	// Admin routes (require ADMIN role)
	admin := api.Group("/admin/assets", middleware.AuthMiddleware(cfg.JWTSecret), middleware.RoleMiddleware("ADMIN"))
	admin.Get("", adminHandler.ListAllAssets)
	admin.Delete("/:id", adminHandler.TakedownAsset)

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