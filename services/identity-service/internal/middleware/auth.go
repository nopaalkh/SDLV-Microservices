package middleware

import (
	"net/http"
	"strings"

	"identity-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware handles JWT authentication
func AuthMiddleware(authUsecase usecase.AuthUsecase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required",
			})
		}

		// Check Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		tokenString := parts[1]

		// Validate token
		claims, err := authUsecase.ValidateToken(c.Context(), tokenString)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Set user information in context
		c.Locals("userID", claims.UserID)
		c.Locals("userRole", string(claims.Role))

		// Continue to the next handler
		return c.Next()
	}
}

// RoleMiddleware checks if the user has the required role
func RoleMiddleware(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user role from context
		userRole, ok := c.Locals("userRole").(string)
		if !ok {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error": "Role information not available",
			})
		}

		// Check if user has the required role
		if userRole != requiredRole {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions",
			})
		}

		// Continue to the next handler
		return c.Next()
	}
}