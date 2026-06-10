package handler

import (
	"net/http"

	"catalog-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// AdminAssetHandler handles admin asset operations
type AdminAssetHandler struct {
	assetUsecase usecase.AssetUsecase
}

// NewAdminAssetHandler creates a new AdminAssetHandler
func NewAdminAssetHandler(assetUsecase usecase.AssetUsecase) *AdminAssetHandler {
	return &AdminAssetHandler{assetUsecase: assetUsecase}
}

// ListAllAssets returns all assets for admin moderation
func (h *AdminAssetHandler) ListAllAssets(c *fiber.Ctx) error {
	assets, err := h.assetUsecase.GetAllAssets(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list assets",
		})
	}

	var resp ListAssetsResponse
	for _, asset := range assets {
		resp.Assets = append(resp.Assets, GetAssetResponse{
			ID:          asset.ID,
			Title:       asset.Title,
			Description: asset.Description,
			Price:       asset.Price,
			AIToolUsed:  asset.AIToolUsed,
			Resolution:  asset.Resolution,
			PreviewURL:  asset.PreviewURL,
			CreatedAt:   asset.CreatedAt,
			UpdatedAt:   asset.UpdatedAt,
			CreatedBy:   asset.CreatedBy,
		})
	}

	return c.JSON(resp)
}

// TakedownAsset forcibly removes an asset (admin moderation)
func (h *AdminAssetHandler) TakedownAsset(c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.assetUsecase.AdminDeleteAsset(c.Context(), id)
	if err != nil {
		if err == usecase.ErrAssetNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Asset not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to takedown asset",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Asset has been taken down",
	})
}
