package handler

import (
	"net/http"
	"strings"

	"catalog-service/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type ReviewHandler struct {
	reviewUsecase usecase.ReviewUsecase
}

func NewReviewHandler(uc usecase.ReviewUsecase) *ReviewHandler {
	return &ReviewHandler{reviewUsecase: uc}
}

type CreateReviewRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

func (h *ReviewHandler) CreateReview(c *fiber.Ctx) error {
	assetID := c.Params("assetId")
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req CreateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	rev, err := h.reviewUsecase.CreateReview(c.Context(), assetID, userID, req.Rating, req.Comment)
	if err != nil {
		if err == usecase.ErrNotPurchased {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "unique constraint") {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Anda sudah pernah mengulas aset ini."})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(rev)
}

func (h *ReviewHandler) GetReviews(c *fiber.Ctx) error {
	assetID := c.Params("assetId")

	reviews, err := h.reviewUsecase.GetReviewsByAssetID(c.Context(), assetID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get reviews"})
	}

	return c.JSON(fiber.Map{"reviews": reviews})
}
