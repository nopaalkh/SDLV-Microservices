package handler

import (
	"net/http"
	"time"

	"identity-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// AdminHandler handles admin-related HTTP requests
type AdminHandler struct {
	authUsecase usecase.AuthUsecase
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(authUsecase usecase.AuthUsecase) *AdminHandler {
	return &AdminHandler{authUsecase: authUsecase}
}

// UserListItem represents a user in the admin list
type UserListItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// ListUsersResponse represents the response for listing all users
type ListUsersResponse struct {
	Users []UserListItem `json:"users"`
	Total int            `json:"total"`
}

// ListUsers returns all users (admin only)
func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	users, err := h.authUsecase.GetAllUsers(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve users",
		})
	}

	var resp ListUsersResponse
	for _, user := range users {
		resp.Users = append(resp.Users, UserListItem{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      string(user.Role),
			Status:    string(user.Status),
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		})
	}
	resp.Total = len(users)

	return c.JSON(resp)
}

// SuspendUser suspends a user account (admin only)
func (h *AdminHandler) SuspendUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	err := h.authUsecase.SuspendUser(c.Context(), userID)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User has been suspended",
	})
}

// UnsuspendUser reactivates a suspended user (admin only)
func (h *AdminHandler) UnsuspendUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	err := h.authUsecase.UnsuspendUser(c.Context(), userID)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User has been reactivated",
	})
}

// DeleteUser deletes a user account (admin only)
func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	err := h.authUsecase.DeleteUser(c.Context(), userID)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User has been deleted",
	})
}
