package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"order-service/internal/repository"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go/snap"
)

// Custom errors
var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidOrderData  = errors.New("invalid order data")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrPaymentFailed     = errors.New("payment failed")
	ErrAssetNotFound     = errors.New("asset not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

// OrderUsecase defines the interface for order use cases
type OrderUsecase interface {
	CreateOrder(ctx context.Context, userID string, assetID string, quantity int) (*repository.Order, string, error)
	GetOrderByID(ctx context.Context, userID string, orderID string) (*repository.Order, error)
	GetUserOrders(ctx context.Context, userID string) ([]*repository.Order, error)
	ProcessPayment(ctx context.Context, userID string, orderID string) error
	MockPayment(ctx context.Context, orderID string) error
	GetAllOrders(ctx context.Context) ([]*repository.Order, error)
	GetCreatorRevenue(ctx context.Context, creatorID string) (float64, error)
	GetPlatformRevenue(ctx context.Context) (float64, error)
}

// orderUsecase implements OrderUsecase
type orderUsecase struct {
	orderRepo      repository.OrderRepository
	assetService   AssetService
	emailService   EmailService
	paymentService PaymentService
	identityClient *IdentityClient
	jwtSecret      string
}

// AssetService defines the interface for interacting with the Catalog Service
type AssetService interface {
	GetAssetDetails(ctx context.Context, assetID string) (*AssetResponse, error)
	ValidateAsset(ctx context.Context, assetID string) (bool, error)
	IncrementAssetSales(ctx context.Context, assetID string) error
}

// EmailService defines the interface for sending emails
type EmailService interface {
	SendOrderConfirmation(ctx context.Context, userEmail string, order *repository.Order) error
}

// PaymentService defines the interface for payment operations
type PaymentService interface {
	CreateSnapTransaction(ctx context.Context, order *repository.Order, userEmail string) (interface{}, error)
	ProcessPaymentSuccess(ctx context.Context, orderID string) error
}

// NewOrderUsecase creates a new OrderUsecase
func NewOrderUsecase(orderRepo repository.OrderRepository, assetService AssetService, emailService EmailService, paymentService PaymentService, identityClient *IdentityClient, jwtSecret string) OrderUsecase {
	return &orderUsecase{
		orderRepo:      orderRepo,
		assetService:   assetService,
		emailService:   emailService,
		paymentService: paymentService,
		identityClient: identityClient,
		jwtSecret:      jwtSecret,
	}
}

// CreateOrder creates a new order
func (uc *orderUsecase) CreateOrder(ctx context.Context, userID string, assetID string, quantity int) (*repository.Order, string, error) {
	// Validate order data
	if assetID == "" || quantity <= 0 {
		return nil, "", ErrInvalidOrderData
	}

	// Validate asset with Catalog Service
	valid, err := uc.assetService.ValidateAsset(ctx, assetID)
	if err != nil {
		return nil, "", err
	}
	if !valid {
		return nil, "", ErrAssetNotFound
	}

	// Get asset details
	assetDetails, err := uc.assetService.GetAssetDetails(ctx, assetID)
	if err != nil {
		return nil, "", err
	}

	// Prevent creator from buying their own asset
	if assetDetails.CreatedBy == userID {
		return nil, "", errors.New("tidak dapat membeli aset buatan sendiri")
	}

	// Calculate total price
	totalPrice := assetDetails.Price * float64(quantity)
	platformFee := totalPrice * 0.08
	creatorRevenue := totalPrice - platformFee

	// Create order
	order := &repository.Order{
		ID:             uuid.New().String(),
		UserID:         userID,
		AssetID:        assetID,
		CreatorID:      assetDetails.CreatedBy,
		Quantity:       quantity,
		TotalPrice:     totalPrice,
		PlatformFee:    platformFee,
		CreatorRevenue: creatorRevenue,
		Status:         repository.StatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Save order
	err = uc.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		return nil, "", err
	}

	// Create Snap transaction
	userEmail, err := uc.identityClient.GetUserEmail(ctx, userID)
	if err != nil {
		userEmail = "user@example.com"
	}
	
	snapResp, err := uc.paymentService.CreateSnapTransaction(ctx, order, userEmail)
	if err != nil {
		return nil, "", err
	}

	// snapResp is interface{}, cast to *snap.Response
	importSnap := snapResp.(*snap.Response)

	return order, importSnap.Token, nil
}

// GetOrderByID retrieves an order by ID
func (uc *orderUsecase) GetOrderByID(ctx context.Context, userID string, orderID string) (*repository.Order, error) {
	order, err := uc.orderRepo.FindOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}

	// Check ownership
	if order.UserID != userID {
		return nil, ErrUnauthorized
	}

	return order, nil
}

// GetUserOrders retrieves all orders for a user
func (uc *orderUsecase) GetUserOrders(ctx context.Context, userID string) ([]*repository.Order, error) {
	orders, err := uc.orderRepo.FindOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// ProcessPayment processes payment for an order
func (uc *orderUsecase) ProcessPayment(ctx context.Context, userID string, orderID string) error {
	// Find order
	order, err := uc.orderRepo.FindOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}

	// Check ownership
	if order.UserID != userID {
		return ErrUnauthorized
	}

	// Check order status
	if order.Status != repository.StatusPending {
		return ErrInvalidOrderData
	}

	// Get user email from Identity Service
	var userEmail string
	if uc.identityClient != nil {
		userEmail, err = uc.identityClient.GetUserEmail(ctx, order.UserID)
		if err != nil {
			userEmail = "user@example.com"
			log.Printf("Failed to get user email from identity service: %v", err)
		}
	} else {
		userEmail = "user@example.com"
	}

	// Create Snap transaction with Midtrans
	snapResp, err := uc.paymentService.CreateSnapTransaction(ctx, order, userEmail)
	if err != nil {
		return err
	}

	// In a real implementation, you would return the Snap token to the frontend
	// and let the frontend handle the payment flow
	log.Printf("Created Snap transaction for order %s: %+v", orderID, snapResp)

	return nil
}

// MockPayment simulates payment for testing purposes
func (uc *orderUsecase) MockPayment(ctx context.Context, orderID string) error {
	// Find order
	order, err := uc.orderRepo.FindOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}

	// Check order status
	if order.Status != repository.StatusPending {
		return ErrInvalidOrderData
	}

	// Process payment success using the payment service
	return uc.paymentService.ProcessPaymentSuccess(ctx, orderID)
}

// GetAllOrders retrieves all orders (admin only)
func (uc *orderUsecase) GetAllOrders(ctx context.Context) ([]*repository.Order, error) {
	return uc.orderRepo.GetAllOrders(ctx)
}

// GetCreatorRevenue gets the total revenue for a creator
func (uc *orderUsecase) GetCreatorRevenue(ctx context.Context, creatorID string) (float64, error) {
	orders, err := uc.orderRepo.FindOrdersByCreatorID(ctx, creatorID)
	if err != nil {
		return 0, err
	}

	var totalRevenue float64
	for _, order := range orders {
		if order.Status == repository.StatusPaid {
			totalRevenue += order.CreatorRevenue
		}
	}

	return totalRevenue, nil
}

// GetPlatformRevenue gets the total revenue for the platform
func (uc *orderUsecase) GetPlatformRevenue(ctx context.Context) (float64, error) {
	return uc.orderRepo.GetTotalPlatformRevenue(ctx)
}
