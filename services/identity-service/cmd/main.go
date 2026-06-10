package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"identity-service/config"
	"identity-service/internal/db"
	"identity-service/internal/handler"
	"identity-service/internal/middleware"
	"identity-service/internal/repository"
	"identity-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
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

	// Initialize database (pass admin credentials from config)
	dbConn, err := db.InitDB(cfg.GetDBConnectionString(), cfg.AdminEmail, cfg.AdminPassword)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbConn.Close()

	// Initialize repository
	userRepo := repository.NewPostgreSQLUserRepository(dbConn.DB)
	prRepo := repository.NewPostgreSQLPasswordResetRepository(dbConn.DB)

	// Initialize email service
	emailSvc := usecase.NewEmailService(cfg.SMTPHost, cfg.SMTPPort, "noreply@marketplace.com")

	// Initialize usecase
	authUsecase := usecase.NewAuthUsecase(userRepo, prRepo, emailSvc, cfg.JWTSecret, 24*time.Hour, cfg.FrontendURL)

	// Initialize handler
	authHandler := handler.NewAuthHandler(authUsecase)
	adminHandler := handler.NewAdminHandler(authUsecase)

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
	auth := api.Group("/auth")

	// Auth routes (with rate limiting on login to prevent brute force)
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many login attempts. Please try again later.",
			})
		},
	}), authHandler.Login)
	auth.Post("/forgot-password", authHandler.ForgotPassword)
	auth.Post("/reset-password", authHandler.ResetPassword)

	// Internal routes (protected with API key for inter-service communication)
	internalAPIKey := cfg.InternalAPIKey
	api.Get("/users/:id", func(c *fiber.Ctx) error {
		key := c.Get("X-Internal-API-Key")
		if key != internalAPIKey {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or missing internal API key",
			})
		}
		return c.Next()
	}, authHandler.GetUserByID)

	// Protected routes
	protected := auth.Group("", middleware.AuthMiddleware(authUsecase))
	protected.Get("/me", authHandler.GetUser)
	protected.Put("/password", authHandler.ChangePassword)

	// Admin routes (require ADMIN role)
	admin := api.Group("/admin", middleware.AuthMiddleware(authUsecase), middleware.RoleMiddleware("ADMIN"))
	admin.Get("/users", adminHandler.ListUsers)
	admin.Put("/users/:id/suspend", adminHandler.SuspendUser)
	admin.Put("/users/:id/unsuspend", adminHandler.UnsuspendUser)
	admin.Delete("/users/:id", adminHandler.DeleteUser)

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