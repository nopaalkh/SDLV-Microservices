package usecase

import (
	"context"
	"errors"
	"time"

	"identity-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// generateUUID generates a new UUID
func generateUUID() string {
	return uuid.New().String()
}

// Custom errors
var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserSuspended      = errors.New("account is suspended")
)

// AuthClaims represents JWT claims with user information
type AuthClaims struct {
	UserID string           `json:"user_id"`
	Email  string           `json:"email"`
	Role   repository.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// AuthUsecase defines the interface for authentication use cases
type AuthUsecase interface {
	Register(ctx context.Context, name, email, password string, role repository.UserRole) (*repository.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	ValidateToken(ctx context.Context, tokenString string) (*AuthClaims, error)
	GetUserByID(ctx context.Context, userID string) (*repository.User, error)
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	GetAllUsers(ctx context.Context) ([]*repository.User, error)
	SuspendUser(ctx context.Context, userID string) error
	UnsuspendUser(ctx context.Context, userID string) error
	DeleteUser(ctx context.Context, userID string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

// authUsecase implements AuthUsecase
type authUsecase struct {
	userRepo       repository.UserRepository
	prRepo         repository.PasswordResetRepository
	emailSvc       EmailService
	jwtSecret      string
	tokenExpiration time.Duration
	frontendURL    string
}

// NewAuthUsecase creates a new AuthUsecase
func NewAuthUsecase(userRepo repository.UserRepository, prRepo repository.PasswordResetRepository, emailSvc EmailService, jwtSecret string, tokenExpiration time.Duration, frontendURL string) AuthUsecase {
	return &authUsecase{
		userRepo:       userRepo,
		prRepo:         prRepo,
		emailSvc:       emailSvc,
		jwtSecret:      jwtSecret,
		tokenExpiration: tokenExpiration,
		frontendURL:    frontendURL,
	}
}

// Register creates a new user
func (uc *authUsecase) Register(ctx context.Context, name, email, password string, role repository.UserRole) (*repository.User, error) {
	// Check if user already exists
	existingUser, err := uc.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password with bcrypt
	hashedPw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashedPassword := string(hashedPw)

	// Create user
	user := &repository.User{
		ID:        generateUUID(),
		Name:      name,
		Email:     email,
		Password:  hashedPassword,
		Role:      role,
		Status:    repository.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save user
	err = uc.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (uc *authUsecase) Login(ctx context.Context, email, password string) (string, error) {
	// Find user by email
	user, err := uc.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	// Check if account is suspended
	if user.Status == repository.StatusSuspended {
		return "", ErrUserSuspended
	}

	// Verify password with bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := uc.generateToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ValidateToken validates a JWT token and returns the claims
func (uc *authUsecase) ValidateToken(ctx context.Context, tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(uc.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetUserByID retrieves a user by ID
func (uc *authUsecase) GetUserByID(ctx context.Context, userID string) (*repository.User, error) {
	user, err := uc.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetAllUsers retrieves all users (admin only)
func (uc *authUsecase) GetAllUsers(ctx context.Context) ([]*repository.User, error) {
	return uc.userRepo.GetAllUsers(ctx)
}

// SuspendUser suspends a user account (admin only)
func (uc *authUsecase) SuspendUser(ctx context.Context, userID string) error {
	user, err := uc.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	if user.Role == repository.RoleAdmin {
		return errors.New("cannot suspend an admin account")
	}
	return uc.userRepo.UpdateUserStatus(ctx, userID, repository.StatusSuspended)
}

// UnsuspendUser reactivates a suspended user account (admin only)
func (uc *authUsecase) UnsuspendUser(ctx context.Context, userID string) error {
	user, err := uc.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	return uc.userRepo.UpdateUserStatus(ctx, userID, repository.StatusActive)
}

// DeleteUser deletes a user account (admin only)
func (uc *authUsecase) DeleteUser(ctx context.Context, userID string) error {
	user, err := uc.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	if user.Role == repository.RoleAdmin {
		return errors.New("cannot delete an admin account")
	}
	return uc.userRepo.DeleteUser(ctx, userID)
}

// generateToken generates a JWT token for a user
func (uc *authUsecase) generateToken(user *repository.User) (string, error) {
	claims := &AuthClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(uc.tokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "identity-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(uc.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ChangePassword allows a user to change their password
func (uc *authUsecase) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := uc.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return errors.New("kata sandi saat ini tidak valid")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update user password
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	return uc.userRepo.UpdateUser(ctx, user)
}

// ForgotPassword generates a reset token and sends an email
func (uc *authUsecase) ForgotPassword(ctx context.Context, email string) error {
	user, err := uc.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		// Do not leak whether email exists
		return nil
	}

	token := generateUUID()
	expiresAt := time.Now().Add(1 * time.Hour)

	if err := uc.prRepo.Create(ctx, user.Email, token, expiresAt); err != nil {
		return err
	}

	resetLink := uc.frontendURL + "/reset-password?token=" + token
	return uc.emailSvc.SendPasswordReset(ctx, user.Email, resetLink)
}

// ResetPassword verifies the token and resets the password
func (uc *authUsecase) ResetPassword(ctx context.Context, token, newPassword string) error {
	pr, err := uc.prRepo.GetByToken(ctx, token)
	if err != nil {
		return err
	}
	if pr == nil {
		return errors.New("invalid or expired token")
	}

	if time.Now().After(pr.ExpiresAt) {
		_ = uc.prRepo.DeleteByEmail(ctx, pr.Email)
		return errors.New("token has expired")
	}

	user, err := uc.userRepo.FindUserByEmail(ctx, pr.Email)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := uc.userRepo.UpdateUser(ctx, user); err != nil {
		return err
	}

	// Invalidate the token
	_ = uc.prRepo.DeleteByEmail(ctx, pr.Email)
	return nil
}
