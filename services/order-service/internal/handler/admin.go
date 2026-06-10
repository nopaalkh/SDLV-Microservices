package handler

import (
	"net/http"

	"order-service/internal/repository"
	"order-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// AdminOrderHandler handles admin order operations
type AdminOrderHandler struct {
	orderUsecase usecase.OrderUsecase
}

// NewAdminOrderHandler creates a new AdminOrderHandler
func NewAdminOrderHandler(orderUsecase usecase.OrderUsecase) *AdminOrderHandler {
	return &AdminOrderHandler{orderUsecase: orderUsecase}
}

// ListAllOrders returns all orders for admin monitoring
func (h *AdminOrderHandler) ListAllOrders(c *fiber.Ctx) error {
	orders, err := h.orderUsecase.GetAllOrders(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list orders",
		})
	}

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
		if order.Status == repository.StatusPaid {
			orderResp.SecureDownloadToken = order.SecureDownloadToken
		}
		resp.Orders = append(resp.Orders, orderResp)
	}

	return c.JSON(resp)
}

// GetPlatformRevenue returns the total revenue for the platform
func (h *AdminOrderHandler) GetPlatformRevenue(c *fiber.Ctx) error {
	revenue, err := h.orderUsecase.GetPlatformRevenue(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate platform revenue",
		})
	}

	return c.JSON(fiber.Map{
		"platform_revenue": revenue,
		"currency":         "IDR",
	})
}
