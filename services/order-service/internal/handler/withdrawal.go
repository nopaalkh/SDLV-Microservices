package handler

import (
	"net/http"

	"order-service/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type WithdrawalHandler struct {
	withdrawalUsecase usecase.WithdrawalUsecase
}

func NewWithdrawalHandler(uc usecase.WithdrawalUsecase) *WithdrawalHandler {
	return &WithdrawalHandler{withdrawalUsecase: uc}
}

type RequestWithdrawalBody struct {
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	BankCode      string  `json:"bank_code" validate:"required"`
	AccountNumber string  `json:"account_number" validate:"required"`
}

func (h *WithdrawalHandler) RequestWithdrawal(c *fiber.Ctx) error {
	creatorID, ok := c.Locals("userID").(string)
	if !ok || creatorID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	role, ok := c.Locals("userRole").(string)
	if !ok || role != "CREATOR" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Only creators can request withdrawals"})
	}

	var req RequestWithdrawalBody
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	w, err := h.withdrawalUsecase.RequestWithdrawal(c.Context(), creatorID, req.Amount, req.BankCode, req.AccountNumber)
	if err != nil {
		if err == usecase.ErrInsufficientBalance {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to request withdrawal"})
	}

	return c.Status(http.StatusCreated).JSON(w)
}

func (h *WithdrawalHandler) GetMyWithdrawals(c *fiber.Ctx) error {
	creatorID, ok := c.Locals("userID").(string)
	if !ok || creatorID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	withdrawals, err := h.withdrawalUsecase.GetCreatorWithdrawals(c.Context(), creatorID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get withdrawals"})
	}

	return c.JSON(fiber.Map{"withdrawals": withdrawals})
}

func (h *WithdrawalHandler) GetAllWithdrawals(c *fiber.Ctx) error {
	withdrawals, err := h.withdrawalUsecase.GetAllWithdrawals(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get withdrawals"})
	}

	return c.JSON(fiber.Map{"withdrawals": withdrawals})
}

func (h *WithdrawalHandler) ApproveWithdrawal(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.withdrawalUsecase.ApproveWithdrawal(c.Context(), id)
	if err != nil {
		if err == usecase.ErrWithdrawalNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "withdrawal approved"})
}

func (h *WithdrawalHandler) RejectWithdrawal(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.withdrawalUsecase.RejectWithdrawal(c.Context(), id)
	if err != nil {
		if err == usecase.ErrWithdrawalNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "withdrawal rejected"})
}
