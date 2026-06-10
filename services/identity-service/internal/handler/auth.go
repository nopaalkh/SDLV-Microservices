package handler

import (
	"net/http"
	"strings"
	"time"

	"identity-service/internal/repository"
	"identity-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authUsecase usecase.AuthUsecase
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authUsecase usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=BUYER CREATOR ADMIN"`
}

// RegisterResponse represents a user registration response
type RegisterResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.Role) == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Name, email, password, and role are required",
		})
	}

	// Validate email format (basic check)
	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid email format",
		})
	}

	// Validate password length
	if len(req.Password) < 6 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 6 characters",
		})
	}

	// Validate role (ADMIN cannot self-register)
	role := repository.UserRole(req.Role)
	if role != repository.RoleBuyer && role != repository.RoleCreator {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid role. Must be BUYER or CREATOR",
		})
	}

	// Register user
	user, err := h.authUsecase.Register(c.Context(), req.Name, req.Email, req.Password, role)
	if err != nil {
		if err == usecase.ErrUserAlreadyExists {
			return c.Status(http.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to register user",
		})
	}

	// Prepare response
	resp := RegisterResponse{
		ID:     user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Role:   string(user.Role),
		Status: string(user.Status),
	}

	return c.Status(http.StatusCreated).JSON(resp)
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a user login response
type LoginResponse struct {
	Token string `json:"token"`
}

// Login handles user login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Login user
	token, err := h.authUsecase.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		if err == usecase.ErrInvalidCredentials {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if err == usecase.ErrUserSuspended {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error": "Account is suspended. Please contact admin.",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to login",
		})
	}

	// Prepare response
	resp := LoginResponse{Token: token}

	return c.JSON(resp)
}

// GetUserResponse represents a user information response
type GetUserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// GetUser handles retrieving user information
func (h *AuthHandler) GetUser(c *fiber.Ctx) error {
	// Get user ID from context (set by middleware)
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get user
	user, err := h.authUsecase.GetUserByID(c.Context(), userID)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	// Prepare response
	resp := GetUserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      string(user.Role),
		Status:    string(user.Status),
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}

	return c.JSON(resp)
}

// GetUserByID handles retrieving user information by ID (for internal service use)
func (h *AuthHandler) GetUserByID(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	// Get user
	user, err := h.authUsecase.GetUserByID(c.Context(), userID)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	// Prepare response
	resp := GetUserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      string(user.Role),
		Status:    string(user.Status),
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}

	return c.JSON(resp)
}

// ChangePasswordRequest represents a request to change password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	// Get user ID from context (set by middleware)
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.NewPassword) < 6 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "New password must be at least 6 characters",
		})
	}

	if err := h.authUsecase.ChangePassword(c.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		if err.Error() == "kata sandi saat ini tidak valid" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if err == usecase.ErrUserNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to change password",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password changed successfully",
	})
}

// ForgotPasswordRequest represents a request to reset password
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ForgotPassword handles sending a password reset email
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Email is required",
		})
	}

	// Always return 200 OK to prevent email enumeration
	if err := h.authUsecase.ForgotPassword(c.Context(), req.Email); err != nil {
		// Log error but don't expose it to user
		return c.JSON(fiber.Map{
			"message": "If that email is registered, a password reset link has been sent.",
		})
	}

	return c.JSON(fiber.Map{
		"message": "If that email is registered, a password reset link has been sent.",
	})
}

// ResetPasswordRequest represents a request to set a new password
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// ResetPassword handles resetting the password
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Token == "" || req.NewPassword == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Token and new_password are required",
		})
	}

	if len(req.NewPassword) < 6 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "New password must be at least 6 characters",
		})
	}

	if err := h.authUsecase.ResetPassword(c.Context(), req.Token, req.NewPassword); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password has been successfully reset",
	})
}