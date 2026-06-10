package handler

import (
	"net/http"
	"time"
	"crypto/sha512"
	"encoding/hex"
	"os"

	"order-service/internal/repository"
	"order-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	orderUsecase usecase.OrderUsecase
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(orderUsecase usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{orderUsecase: orderUsecase}
}

// CreateOrderRequest represents a request to create an order
type CreateOrderRequest struct {
	AssetID  string `json:"asset_id" validate:"required"`
	Quantity int    `json:"quantity" validate:"required,gt=0"`
}

// CreateOrderResponse represents a response for order creation
type CreateOrderResponse struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	AssetID        string    `json:"asset_id"`
	Quantity       int       `json:"quantity"`
	TotalPrice     float64   `json:"total_price"`
	PlatformFee    float64   `json:"platform_fee"`
	CreatorRevenue float64   `json:"creator_revenue"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	SnapToken      string    `json:"snap_token,omitempty"`
}

// CreateOrder handles order creation
func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	// Get user ID from context (set by middleware)
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create order
	order, snapToken, err := h.orderUsecase.CreateOrder(c.Context(), userID, req.AssetID, req.Quantity)
	if err != nil {
		if err == usecase.ErrInvalidOrderData {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if err == usecase.ErrAssetNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create order",
		})
	}

	// Prepare response
	resp := CreateOrderResponse{
		ID:             order.ID,
		UserID:         order.UserID,
		AssetID:        order.AssetID,
		Quantity:       order.Quantity,
		TotalPrice:     order.TotalPrice,
		PlatformFee:    order.PlatformFee,
		CreatorRevenue: order.CreatorRevenue,
		Status:         string(order.Status),
		CreatedAt:      order.CreatedAt,
		SnapToken:      snapToken,
	}

	return c.Status(http.StatusCreated).JSON(resp)
}

// GetOrderResponse represents a response for getting an order
type GetOrderResponse struct {
	ID                  string    `json:"id"`
	UserID              string    `json:"user_id"`
	AssetID             string    `json:"asset_id"`
	Quantity            int       `json:"quantity"`
	TotalPrice          float64   `json:"total_price"`
	PlatformFee         float64   `json:"platform_fee"`
	CreatorRevenue      float64   `json:"creator_revenue"`
	Status              string    `json:"status"`
	SecureDownloadToken string    `json:"secure_download_token,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// GetOrder handles getting an order by ID
func (h *OrderHandler) GetOrder(c *fiber.Ctx) error {
	// Get user ID from context (set by middleware)
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}
	orderID := c.Params("id")

	order, err := h.orderUsecase.GetOrderByID(c.Context(), userID, orderID)
	if err != nil {
		if err == usecase.ErrOrderNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if err == usecase.ErrUnauthorized {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get order",
		})
	}

	// Prepare response
	resp := GetOrderResponse{
		ID:             order.ID,
		UserID:         order.UserID,
		AssetID:        order.AssetID,
		Quantity:       order.Quantity,
		TotalPrice:     order.TotalPrice,
		PlatformFee:    order.PlatformFee,
		CreatorRevenue: order.CreatorRevenue,
		Status:         string(order.Status),
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}

	// Only include download token for paid orders
	if order.Status == repository.StatusPaid {
		resp.SecureDownloadToken = order.SecureDownloadToken
	}

	return c.JSON(resp)
}

// ListOrdersResponse represents a response for listing orders
type ListOrdersResponse struct {
	Orders []GetOrderResponse `json:"orders"`
}

// ListOrders handles listing orders for the current user
func (h *OrderHandler) ListOrders(c *fiber.Ctx) error {
	// Get user ID from context (set by middleware)
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	orders, err := h.orderUsecase.GetUserOrders(c.Context(), userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list orders",
		})
	}

	// Prepare response
	var resp ListOrdersResponse
	for _, order := range orders {
		orderResp := GetOrderResponse{
			ID:             order.ID,
			UserID:         order.UserID,
			AssetID:        order.AssetID,
			Quantity:       order.Quantity,
			TotalPrice:     order.TotalPrice,
			PlatformFee:    order.PlatformFee,
			CreatorRevenue: order.CreatorRevenue,
			Status:         string(order.Status),
			CreatedAt:      order.CreatedAt,
			UpdatedAt:      order.UpdatedAt,
		}
		// Only include download token for paid orders
		if order.Status == repository.StatusPaid {
			orderResp.SecureDownloadToken = order.SecureDownloadToken
		}
		resp.Orders = append(resp.Orders, orderResp)
	}

	return c.JSON(resp)
}

// ProcessPayment handles order payment
func (h *OrderHandler) ProcessPayment(c *fiber.Ctx) error {
	// Get user ID from context (set by middleware)
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}
	orderID := c.Params("id")

	err := h.orderUsecase.ProcessPayment(c.Context(), userID, orderID)
	if err != nil {
		if err == usecase.ErrOrderNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if err == usecase.ErrUnauthorized {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if err == usecase.ErrInvalidOrderData {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process payment",
		})
	}

	return c.JSON(fiber.Map{"status": "payment processed successfully"})
}

// MockPayment handles mock payment for testing
func (h *OrderHandler) MockPayment(c *fiber.Ctx) error {
	orderID := c.Params("id")

	err := h.orderUsecase.MockPayment(c.Context(), orderID)
	if err != nil {
		if err == usecase.ErrOrderNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if err == usecase.ErrInvalidOrderData {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process mock payment",
		})
	}

	return c.JSON(fiber.Map{"status": "mock payment processed successfully"})
}

// GetRevenue handles fetching the total revenue for a creator
func (h *OrderHandler) GetRevenue(c *fiber.Ctx) error {
	// Get user ID (creator ID) from context
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Verify role is CREATOR
	role, ok := c.Locals("userRole").(string)
	if !ok || role != "CREATOR" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Only creators can view revenue",
		})
	}

	revenue, err := h.orderUsecase.GetCreatorRevenue(c.Context(), userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate revenue",
		})
	}

	return c.JSON(fiber.Map{
		"creator_id": userID,
		"revenue":    revenue,
		"currency":   "IDR",
	})
}

// CheckPurchase checks if a user has purchased an asset
func (h *OrderHandler) CheckPurchase(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	assetID := c.Query("asset_id")

	if userID == "" || assetID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "user_id and asset_id are required"})
	}

	orders, err := h.orderUsecase.GetUserOrders(c.Context(), userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get user orders"})
	}

	hasPurchased := false
	for _, order := range orders {
		if order.AssetID == assetID && order.Status == "PAID" {
			hasPurchased = true
			break
		}
	}

	return c.JSON(fiber.Map{"has_purchased": hasPurchased})
}

// MidtransWebhook handles notifications from Midtrans
func (h *OrderHandler) MidtransWebhook(c *fiber.Ctx) error {
	var notificationPayload map[string]interface{}
	if err := c.BodyParser(&notificationPayload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse notification payload",
		})
	}

	orderId, ok := notificationPayload["order_id"].(string)
	if !ok {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "order_id not found in payload",
		})
	}

	statusCode, ok := notificationPayload["status_code"].(string)
	if !ok {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "status_code not found in payload",
		})
	}

	grossAmount, ok := notificationPayload["gross_amount"].(string)
	if !ok {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "gross_amount not found in payload",
		})
	}

	signatureKey, ok := notificationPayload["signature_key"].(string)
	if !ok {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "signature_key not found in payload",
		})
	}

	// Verify signature
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	strToHash := orderId + statusCode + grossAmount + serverKey
	hasher := sha512.New()
	hasher.Write([]byte(strToHash))
	calculatedSignature := hex.EncodeToString(hasher.Sum(nil))

	if calculatedSignature != signatureKey {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid signature key",
		})
	}

	transactionStatus, ok := notificationPayload["transaction_status"].(string)
	if !ok {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "transaction_status not found in payload",
		})
	}

	if transactionStatus == "capture" || transactionStatus == "settlement" {
		// Only process if it hasn't been processed yet. `MockPayment` actually handles the internal `ProcessPaymentSuccess`
		err := h.orderUsecase.MockPayment(c.Context(), orderId)
		if err != nil && err.Error() != "order has already been paid" {
			// ignore if already paid
		}
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "ok",
	})
}
