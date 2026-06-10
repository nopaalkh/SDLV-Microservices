package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware handles JWT authentication by validating the token
func AuthMiddleware(jwtSecret string) fiber.Handler {
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
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Extract user information from claims
			userID, ok1 := claims["user_id"].(string)
			userRole, ok2 := claims["role"].(string)

			if !ok1 || !ok2 {
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid token claims",
				})
			}

			// Set user information in context
			c.Locals("userID", userID)
			c.Locals("userRole", userRole)

			// Continue to the next handler
			return c.Next()
		}

		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
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