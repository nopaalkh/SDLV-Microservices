package usecase

import (
	"context"
	"errors"
	"log"
	"order-service/config"
	"order-service/internal/repository"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

// PaymentService interface is defined in order.go

// midtransService implements PaymentService
type midtransService struct {
	config         *config.AppConfig
	orderRepo      repository.OrderRepository
	emailService   EmailService
	identityClient *IdentityClient
	assetService   AssetService
}

// NewPaymentService creates a new PaymentService
func NewPaymentService(config *config.AppConfig, orderRepo repository.OrderRepository, emailService EmailService, identityClient *IdentityClient, assetService AssetService) PaymentService {
	return &midtransService{
		config:         config,
		orderRepo:      orderRepo,
		emailService:   emailService,
		identityClient: identityClient,
		assetService:   assetService,
	}
}

// CreateSnapTransaction creates a Snap transaction for Midtrans payment gateway
func (s *midtransService) CreateSnapTransaction(ctx context.Context, order *repository.Order, userEmail string) (interface{}, error) {
	// Initialize Midtrans client
	midtrans.ClientKey = s.config.MidtransClientKey
	midtrans.ServerKey = s.config.MidtransServerKey
	midtrans.Environment = midtrans.Sandbox
	if s.config.MidtransIsProduction {
		midtrans.Environment = midtrans.Production
	}

	// Create Snap request
	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  order.ID,
			GrossAmt: int64(order.TotalPrice),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			Email: userEmail,
		},
		Items: &[]midtrans.ItemDetails{
			{
				ID:    order.AssetID,
				Price: int64(order.TotalPrice / float64(order.Quantity)),
				Qty:   int32(order.Quantity),
				Name:  "Digital Asset", // This would be fetched from catalog service in a real implementation
			},
		},
	}

	// Create Snap transaction
	snapClient := snap.Client{}
	snapClient.New(midtrans.ServerKey, midtrans.Environment)

	resp, err := snapClient.CreateTransaction(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}


// ProcessPaymentSuccess handles successful payment processing
func (s *midtransService) ProcessPaymentSuccess(ctx context.Context, orderID string) error {
	// Find the order
	order, err := s.orderRepo.FindOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	// Generate download token
	downloadToken := uuid.New().String()

	// Update order status to PAID
	err = s.orderRepo.UpdateOrderPayment(ctx, order.ID, repository.StatusPaid, downloadToken)
	if err != nil {
		return err
	}

	// Get user email from Identity Service
	userEmail, err := s.identityClient.GetUserEmail(ctx, order.UserID)
	if err != nil {
		userEmail = "user@example.com"
		log.Printf("Failed to get user email from identity service: %v", err)
	}

	// Increment asset sales count asynchronously
	go func() {
		ctx := context.Background()
		if err := s.assetService.IncrementAssetSales(ctx, order.AssetID); err != nil {
			log.Printf("Failed to increment asset sales count: %v", err)
		}
	}()

	// Send confirmation email asynchronously
	go func() {
		emailCtx := context.Background()
		_ = s.emailService.SendOrderConfirmation(emailCtx, userEmail, order)
	}()

	return nil
}
