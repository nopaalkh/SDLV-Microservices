package usecase_test

import (
	"context"
	"testing"
	"time"

	"order-service/internal/repository"
	"order-service/internal/usecase"

	"github.com/midtrans/midtrans-go/snap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrderRepository is a mock implementation of OrderRepository
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, order *repository.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) FindOrderByID(ctx context.Context, id string) (*repository.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Order), args.Error(1)
}

func (m *MockOrderRepository) FindOrdersByUserID(ctx context.Context, userID string) ([]*repository.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.Order), args.Error(1)
}

func (m *MockOrderRepository) FindOrdersByCreatorID(ctx context.Context, creatorID string) ([]*repository.Order, error) {
	args := m.Called(ctx, creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.Order), args.Error(1)
}

func (m *MockOrderRepository) GetAllOrders(ctx context.Context) ([]*repository.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.Order), args.Error(1)
}

func (m *MockOrderRepository) GetTotalPlatformRevenue(ctx context.Context) (float64, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockOrderRepository) UpdateOrderStatus(ctx context.Context, id string, status repository.OrderStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateOrderPayment(ctx context.Context, id string, status repository.OrderStatus, downloadToken string) error {
	args := m.Called(ctx, id, status, downloadToken)
	return args.Error(0)
}

// MockAssetService is a mock implementation of AssetService
type MockAssetService struct {
	mock.Mock
}

func (m *MockAssetService) GetAssetDetails(ctx context.Context, assetID string) (*usecase.AssetResponse, error) {
	args := m.Called(ctx, assetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.AssetResponse), args.Error(1)
}

func (m *MockAssetService) ValidateAsset(ctx context.Context, assetID string) (bool, error) {
	args := m.Called(ctx, assetID)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockAssetService) IncrementAssetSales(ctx context.Context, assetID string) error {
	args := m.Called(ctx, assetID)
	return args.Error(0)
}

// MockEmailService is a mock implementation of EmailService
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendOrderConfirmation(ctx context.Context, userEmail string, order *repository.Order) error {
	args := m.Called(ctx, userEmail, order)
	return args.Error(0)
}

// MockPaymentService is a mock implementation of PaymentService
type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) CreateSnapTransaction(ctx context.Context, order *repository.Order, userEmail string) (interface{}, error) {
	args := m.Called(ctx, order, userEmail)
	return args.Get(0).(*snap.Response), args.Error(1)
}

func (m *MockPaymentService) ProcessPaymentSuccess(ctx context.Context, orderID string) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func TestCreateOrder_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockOrderRepository)
	mockAssetService := new(MockAssetService)
	mockEmailService := new(MockEmailService)
	mockPaymentService := &MockPaymentService{}

	uc := usecase.NewOrderUsecase(mockRepo, mockAssetService, mockEmailService, mockPaymentService, nil, "test-secret")

	// Mock expectations
	mockAssetService.On("ValidateAsset", mock.Anything, "asset-123").Return(true, nil)
	mockAssetService.On("GetAssetDetails", mock.Anything, "asset-123").Return(&usecase.AssetResponse{
		ID:        "asset-123",
		Price:     10.99,
		CreatedBy: "creator-123",
	}, nil)
	mockRepo.On("CreateOrder", mock.Anything, mock.AnythingOfType("*repository.Order")).Return(nil)

	// Test
	order, err := uc.CreateOrder(context.Background(), "user-123", "asset-123", 2)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, "user-123", order.UserID)
	assert.Equal(t, "asset-123", order.AssetID)
	assert.Equal(t, 2, order.Quantity)
	assert.Equal(t, 21.98, order.TotalPrice)
	assert.Equal(t, repository.StatusPending, order.Status)

	// Verify mocks
	mockAssetService.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestCreateOrder_InvalidData(t *testing.T) {
	// Setup
	mockRepo := new(MockOrderRepository)
	mockAssetService := new(MockAssetService)
	mockEmailService := new(MockEmailService)
	mockPaymentService := &MockPaymentService{}

	uc := usecase.NewOrderUsecase(mockRepo, mockAssetService, mockEmailService, mockPaymentService, nil, "test-secret")

	// Test invalid asset ID
	order, err := uc.CreateOrder(context.Background(), "user-123", "", 2)
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Equal(t, usecase.ErrInvalidOrderData, err)

	// Test invalid quantity
	order, err = uc.CreateOrder(context.Background(), "user-123", "asset-123", 0)
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Equal(t, usecase.ErrInvalidOrderData, err)
}

func TestCreateOrder_AssetNotFound(t *testing.T) {
	// Setup
	mockRepo := new(MockOrderRepository)
	mockAssetService := new(MockAssetService)
	mockEmailService := new(MockEmailService)
	mockPaymentService := &MockPaymentService{}

	uc := usecase.NewOrderUsecase(mockRepo, mockAssetService, mockEmailService, mockPaymentService, nil, "test-secret")

	// Mock expectations
	mockAssetService.On("ValidateAsset", mock.Anything, "asset-123").Return(false, nil)

	// Test
	order, err := uc.CreateOrder(context.Background(), "user-123", "asset-123", 2)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Equal(t, usecase.ErrAssetNotFound, err)

	// Verify mocks
	mockAssetService.AssertExpectations(t)
}

func TestGetOrderByID_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockOrderRepository)
	mockAssetService := new(MockAssetService)
	mockEmailService := new(MockEmailService)
	mockPaymentService := &MockPaymentService{}

	uc := usecase.NewOrderUsecase(mockRepo, mockAssetService, mockEmailService, mockPaymentService, nil, "test-secret")

	// Mock data
	testOrder := &repository.Order{
		ID:         "order-123",
		UserID:     "user-123",
		AssetID:    "asset-123",
		Quantity:   2,
		TotalPrice: 21.98,
		Status:     repository.StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Mock expectations
	mockRepo.On("FindOrderByID", mock.Anything, "order-123").Return(testOrder, nil)

	// Test
	order, err := uc.GetOrderByID(context.Background(), "user-123", "order-123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, "order-123", order.ID)
	assert.Equal(t, "user-123", order.UserID)

	// Verify mocks
	mockRepo.AssertExpectations(t)
}

func TestGetOrderByID_NotFound(t *testing.T) {
	// Setup
	mockRepo := new(MockOrderRepository)
	mockAssetService := new(MockAssetService)
	mockEmailService := new(MockEmailService)

	mockPaymentService := &MockPaymentService{}
	uc := usecase.NewOrderUsecase(mockRepo, mockAssetService, mockEmailService, mockPaymentService, nil, "test-secret")

	// Mock expectations
	mockRepo.On("FindOrderByID", mock.Anything, "order-123").Return(nil, nil)

	// Test
	order, err := uc.GetOrderByID(context.Background(), "user-123", "order-123")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Equal(t, usecase.ErrOrderNotFound, err)

	// Verify mocks
	mockRepo.AssertExpectations(t)
}

func TestGetOrderByID_Unauthorized(t *testing.T) {
	// Setup
	mockRepo := new(MockOrderRepository)
	mockAssetService := new(MockAssetService)
	mockEmailService := new(MockEmailService)
	mockPaymentService := &MockPaymentService{}

	uc := usecase.NewOrderUsecase(mockRepo, mockAssetService, mockEmailService, mockPaymentService, nil, "test-secret")

	// Mock data
	testOrder := &repository.Order{
		ID:         "order-123",
		UserID:     "user-456", // Different user
		AssetID:    "asset-123",
		Quantity:   2,
		TotalPrice: 21.98,
		Status:     repository.StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Mock expectations
	mockRepo.On("FindOrderByID", mock.Anything, "order-123").Return(testOrder, nil)

	// Test
	order, err := uc.GetOrderByID(context.Background(), "user-123", "order-123")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Equal(t, usecase.ErrUnauthorized, err)

	// Verify mocks
	mockRepo.AssertExpectations(t)
}

func TestProcessPayment_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockOrderRepository)
	mockAssetService := new(MockAssetService)
	mockEmailService := new(MockEmailService)
	mockPaymentService := &MockPaymentService{}

	uc := usecase.NewOrderUsecase(mockRepo, mockAssetService, mockEmailService, mockPaymentService, nil, "test-secret")

	// Mock data
	testOrder := &repository.Order{
		ID:         "order-123",
		UserID:     "user-123",
		AssetID:    "asset-123",
		Quantity:   2,
		TotalPrice: 21.98,
		Status:     repository.StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Mock expectations
	mockRepo.On("FindOrderByID", mock.Anything, "order-123").Return(testOrder, nil)
	mockPaymentService.On("CreateSnapTransaction", mock.Anything, mock.AnythingOfType("*repository.Order"), mock.AnythingOfType("string")).Return(&snap.Response{Token: "test-token"}, nil)

	// Test
	err := uc.ProcessPayment(context.Background(), "user-123", "order-123")
	assert.NoError(t, err)
	mockPaymentService.AssertExpectations(t)

	// Assert
	assert.NoError(t, err)

	// Verify mocks
	mockRepo.AssertExpectations(t)
	mockPaymentService.AssertExpectations(t)
}

func TestProcessPayment_OrderNotFound(t *testing.T) {
	// Setup
	mockRepo := new(MockOrderRepository)
	mockAssetService := new(MockAssetService)
	mockEmailService := new(MockEmailService)
	mockPaymentService := &MockPaymentService{}

	uc := usecase.NewOrderUsecase(mockRepo, mockAssetService, mockEmailService, mockPaymentService, nil, "test-secret")

	// Mock expectations
	mockRepo.On("FindOrderByID", mock.Anything, "order-123").Return(nil, nil)

	// Test
	err := uc.ProcessPayment(context.Background(), "user-123", "order-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, usecase.ErrOrderNotFound, err)

	// Verify mocks
	mockRepo.AssertExpectations(t)
}

func TestProcessPayment_Unauthorized(t *testing.T) {
	// Setup
	mockRepo := new(MockOrderRepository)
	mockAssetService := new(MockAssetService)
	mockEmailService := new(MockEmailService)
	mockPaymentService := &MockPaymentService{}

	uc := usecase.NewOrderUsecase(mockRepo, mockAssetService, mockEmailService, mockPaymentService, nil, "test-secret")

	// Mock data
	testOrder := &repository.Order{
		ID:         "order-123",
		UserID:     "user-456", // Different user
		AssetID:    "asset-123",
		Quantity:   2,
		TotalPrice: 21.98,
		Status:     repository.StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Mock expectations
	mockRepo.On("FindOrderByID", mock.Anything, "order-123").Return(testOrder, nil)

	// Test
	err := uc.ProcessPayment(context.Background(), "user-123", "order-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, usecase.ErrUnauthorized, err)

	// Verify mocks
	mockRepo.AssertExpectations(t)
}

func TestProcessPayment_InvalidStatus(t *testing.T) {
	// Setup
	mockRepo := new(MockOrderRepository)
	mockAssetService := new(MockAssetService)
	mockEmailService := new(MockEmailService)
	mockPaymentService := &MockPaymentService{}

	uc := usecase.NewOrderUsecase(mockRepo, mockAssetService, mockEmailService, mockPaymentService, nil, "test-secret")

	// Mock data
	testOrder := &repository.Order{
		ID:         "order-123",
		UserID:     "user-123",
		AssetID:    "asset-123",
		Quantity:   2,
		TotalPrice: 21.98,
		Status:     repository.StatusPaid, // Already paid
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Mock expectations
	mockRepo.On("FindOrderByID", mock.Anything, "order-123").Return(testOrder, nil)

	// Test
	err := uc.ProcessPayment(context.Background(), "user-123", "order-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, usecase.ErrInvalidOrderData, err)

	// Verify mocks
	mockRepo.AssertExpectations(t)
}

func TestMockPayment_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockOrderRepository)
	mockAssetService := new(MockAssetService)
	mockEmailService := new(MockEmailService)
	mockPaymentService := &MockPaymentService{}

	uc := usecase.NewOrderUsecase(mockRepo, mockAssetService, mockEmailService, mockPaymentService, nil, "test-secret")

	// Mock data
	testOrder := &repository.Order{
		ID:         "order-123",
		UserID:     "user-123",
		AssetID:    "asset-123",
		Quantity:   2,
		TotalPrice: 21.98,
		Status:     repository.StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Mock expectations
	mockRepo.On("FindOrderByID", mock.Anything, "order-123").Return(testOrder, nil)
	mockPaymentService.On("ProcessPaymentSuccess", mock.Anything, "order-123").Return(nil)

	// Test
	err := uc.MockPayment(context.Background(), "order-123")

	// Assert
	assert.NoError(t, err)

	// Verify mocks
	mockRepo.AssertExpectations(t)
	mockPaymentService.AssertExpectations(t)
}

func TestMockPayment_OrderNotFound(t *testing.T) {
	// Setup
	mockRepo := new(MockOrderRepository)
	mockAssetService := new(MockAssetService)
	mockEmailService := new(MockEmailService)
	mockPaymentService := &MockPaymentService{}

	uc := usecase.NewOrderUsecase(mockRepo, mockAssetService, mockEmailService, mockPaymentService, nil, "test-secret")

	// Mock expectations
	mockRepo.On("FindOrderByID", mock.Anything, "order-123").Return(nil, nil)

	// Test
	err := uc.MockPayment(context.Background(), "order-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, usecase.ErrOrderNotFound, err)

	// Verify mocks
	mockRepo.AssertExpectations(t)
}

func TestMockPayment_InvalidStatus(t *testing.T) {
	// Setup
	mockRepo := new(MockOrderRepository)
	mockAssetService := new(MockAssetService)
	mockEmailService := new(MockEmailService)
	mockPaymentService := &MockPaymentService{}

	uc := usecase.NewOrderUsecase(mockRepo, mockAssetService, mockEmailService, mockPaymentService, nil, "test-secret")

	// Mock data
	testOrder := &repository.Order{
		ID:         "order-123",
		UserID:     "user-123",
		AssetID:    "asset-123",
		Quantity:   2,
		TotalPrice: 21.98,
		Status:     repository.StatusPaid, // Already paid
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Mock expectations
	mockRepo.On("FindOrderByID", mock.Anything, "order-123").Return(testOrder, nil)

	// Test
	err := uc.MockPayment(context.Background(), "order-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, usecase.ErrInvalidOrderData, err)

	// Verify mocks
	mockRepo.AssertExpectations(t)
}
