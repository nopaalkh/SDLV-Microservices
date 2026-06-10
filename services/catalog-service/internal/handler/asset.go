package handler

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"catalog-service/internal/repository"
	"catalog-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// AssetHandler handles asset-related HTTP requests
type AssetHandler struct {
	assetUsecase usecase.AssetUsecase
	orderClient  usecase.OrderClient
}

// NewAssetHandler creates a new AssetHandler
func NewAssetHandler(assetUsecase usecase.AssetUsecase, orderClient usecase.OrderClient) *AssetHandler {
	return &AssetHandler{assetUsecase: assetUsecase, orderClient: orderClient}
}

// CreateAssetRequest represents a request to create an asset
type CreateAssetRequest struct {
	Title        string  `json:"title" form:"title" validate:"required"`
	Description  string  `json:"description" form:"description"`
	Price        float64 `json:"price" form:"price" validate:"required,gt=0"`
	Category     string  `json:"category" form:"category"`
	AIToolUsed   string  `json:"ai_tool_used" form:"ai_tool_used"`
	Resolution   string  `json:"resolution" form:"resolution"`
	PreviewURL   string  `json:"preview_url" form:"preview_url"`
	StoragePath  string  `json:"storage_path" form:"storage_path"`
}

// CreateAssetResponse represents a response for asset creation
type CreateAssetResponse struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Price        float64   `json:"price"`
	Category     string    `json:"category"`
	AIToolUsed   string    `json:"ai_tool_used"`
	Resolution   string    `json:"resolution"`
	PreviewURL   string    `json:"preview_url"`
	SalesCount   int       `json:"sales_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateAsset handles asset creation
func (h *AssetHandler) CreateAsset(c *fiber.Ctx) error {
	// Get user ID from context (set by middleware)
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req CreateAssetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Handle file upload
	file, err := c.FormFile("file")
	if err == nil {
		// Save file
		os.MkdirAll("./uploads", 0755)
		filename := fmt.Sprintf("./uploads/%d-%s", time.Now().Unix(), file.Filename)
		if err := c.SaveFile(file, filename); err == nil {
			req.StoragePath = filename
		}
	}

	// Ensure title is present (e2e fallback)
	if req.Title == "" {
		req.Title = c.FormValue("title", "Untitled Asset")
	}

	// Create asset
	asset := &repository.Asset{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		AIToolUsed:  req.AIToolUsed,
		Resolution:  req.Resolution,
		PreviewURL:  req.PreviewURL,
		StoragePath: req.StoragePath,
	}

	createdAsset, err := h.assetUsecase.CreateAsset(c.Context(), asset, userID)
	if err != nil {
		if err == usecase.ErrInvalidAssetData {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create asset",
		})
	}

	// Prepare response
	resp := CreateAssetResponse{
		ID:          createdAsset.ID,
		Title:       createdAsset.Title,
		Description: createdAsset.Description,
		Price:       createdAsset.Price,
		Category:    createdAsset.Category,
		AIToolUsed:  createdAsset.AIToolUsed,
		Resolution:  createdAsset.Resolution,
		PreviewURL:  createdAsset.PreviewURL,
		SalesCount:  createdAsset.SalesCount,
		CreatedAt:   createdAsset.CreatedAt,
		UpdatedAt:   createdAsset.UpdatedAt,
	}

	return c.Status(http.StatusCreated).JSON(resp)
}

// GetAssetResponse represents a response for getting an asset
type GetAssetResponse struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Price        float64   `json:"price"`
	Category     string    `json:"category"`
	AIToolUsed   string    `json:"ai_tool_used"`
	Resolution   string    `json:"resolution"`
	PreviewURL   string    `json:"preview_url"`
	SalesCount   int       `json:"sales_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedBy    string    `json:"created_by"`
}

// GetAsset handles getting an asset by ID
func (h *AssetHandler) GetAsset(c *fiber.Ctx) error {
	id := c.Params("id")

	asset, err := h.assetUsecase.GetAssetByID(c.Context(), id)
	if err != nil {
		if err == usecase.ErrAssetNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get asset",
		})
	}

	// Prepare response
	resp := GetAssetResponse{
		ID:          asset.ID,
		Title:       asset.Title,
		Description: asset.Description,
		Price:       asset.Price,
		Category:    asset.Category,
		AIToolUsed:  asset.AIToolUsed,
		Resolution:  asset.Resolution,
		PreviewURL:  asset.PreviewURL,
		SalesCount:  asset.SalesCount,
		CreatedAt:   asset.CreatedAt,
		UpdatedAt:   asset.UpdatedAt,
		CreatedBy:   asset.CreatedBy,
	}

	return c.JSON(resp)
}

// GetDownloadLink handles getting the secure Google Drive link after purchase verification
func (h *AssetHandler) GetDownloadLink(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "User not authenticated"})
	}

	hasPurchased, err := h.orderClient.HasPurchased(userID, id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to verify purchase"})
	}
	if !hasPurchased {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "You must purchase this asset first to get the access link"})
	}

	asset, err := h.assetUsecase.GetAssetByID(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Asset not found"})
	}

	return c.JSON(fiber.Map{"download_url": asset.StoragePath})
}

// ListAssetsResponse represents a response for listing assets
type ListAssetsResponse struct {
	Assets []GetAssetResponse `json:"assets"`
}

// ListAssets handles listing assets by creator
func (h *AssetHandler) ListAssets(c *fiber.Ctx) error {
	// Get user ID from context (set by middleware)
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	assets, err := h.assetUsecase.GetAssetsByCreator(c.Context(), userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list assets",
		})
	}

	// Prepare response
	var resp ListAssetsResponse
	for _, asset := range assets {
		resp.Assets = append(resp.Assets, GetAssetResponse{
			ID:          asset.ID,
			Title:       asset.Title,
			Description: asset.Description,
			Price:       asset.Price,
			Category:    asset.Category,
			AIToolUsed:  asset.AIToolUsed,
			Resolution:  asset.Resolution,
			PreviewURL:  asset.PreviewURL,
			SalesCount:  asset.SalesCount,
			CreatedAt:   asset.CreatedAt,
			UpdatedAt:   asset.UpdatedAt,
			CreatedBy:   asset.CreatedBy,
		})
	}

	return c.JSON(resp)
}

// SearchAssetsRequest represents a request to search assets
type SearchAssetsRequest struct {
	Query    string `json:"query"`
	Category string `json:"category"`
	SortBy   string `json:"sort_by"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

// SearchAssets handles searching assets
func (h *AssetHandler) SearchAssets(c *fiber.Ctx) error {
	var req SearchAssetsRequest
	if err := c.BodyParser(&req); err != nil {
		// If body parser fails, it might be a GET request or empty body. That's fine, we'll use defaults.
	}

	assets, err := h.assetUsecase.SearchAssets(c.Context(), req.Query, req.Category, req.SortBy, req.Limit, req.Offset)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search assets",
		})
	}

	// Prepare response
	var resp ListAssetsResponse
	for _, asset := range assets {
		resp.Assets = append(resp.Assets, GetAssetResponse{
			ID:          asset.ID,
			Title:       asset.Title,
			Description: asset.Description,
			Price:       asset.Price,
			Category:    asset.Category,
			AIToolUsed:  asset.AIToolUsed,
			Resolution:  asset.Resolution,
			PreviewURL:  asset.PreviewURL,
			SalesCount:  asset.SalesCount,
			CreatedAt:   asset.CreatedAt,
			UpdatedAt:   asset.UpdatedAt,
			CreatedBy:   asset.CreatedBy,
		})
	}

	return c.JSON(resp)
}

// UpdateAssetRequest represents a request to update an asset
type UpdateAssetRequest struct {
	Title        string  `json:"title" validate:"required"`
	Description  string  `json:"description"`
	Price        float64 `json:"price" validate:"required,gt=0"`
	Category     string  `json:"category"`
	AIToolUsed   string  `json:"ai_tool_used"`
	Resolution   string  `json:"resolution"`
	PreviewURL   string  `json:"preview_url"`
	StoragePath  string  `json:"storage_path" validate:"required"`
}

// UpdateAsset handles updating an asset
func (h *AssetHandler) UpdateAsset(c *fiber.Ctx) error {
	// Get user ID from context (set by middleware)
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req UpdateAssetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create asset
	asset := &repository.Asset{
		ID:          c.Params("id"),
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		AIToolUsed:  req.AIToolUsed,
		Resolution:  req.Resolution,
		PreviewURL:  req.PreviewURL,
		StoragePath: req.StoragePath,
	}

	updatedAsset, err := h.assetUsecase.UpdateAsset(c.Context(), asset, userID)
	if err != nil {
		if err == usecase.ErrAssetNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if err == usecase.ErrUnauthorized {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if err == usecase.ErrInvalidAssetData {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update asset",
		})
	}

	// Prepare response
	resp := GetAssetResponse{
		ID:          updatedAsset.ID,
		Title:       updatedAsset.Title,
		Description: updatedAsset.Description,
		Price:       updatedAsset.Price,
		AIToolUsed:  updatedAsset.AIToolUsed,
		Resolution:  updatedAsset.Resolution,
		PreviewURL:  updatedAsset.PreviewURL,
		SalesCount:  updatedAsset.SalesCount,
		CreatedAt:   updatedAsset.CreatedAt,
		UpdatedAt:   updatedAsset.UpdatedAt,
		CreatedBy:   updatedAsset.CreatedBy,
	}

	return c.JSON(resp)
}

// DeleteAsset handles deleting an asset
func (h *AssetHandler) DeleteAsset(c *fiber.Ctx) error {
	// Get user ID from context (set by middleware)
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}
	id := c.Params("id")

	err := h.assetUsecase.DeleteAsset(c.Context(), id, userID)
	if err != nil {
		if err == usecase.ErrAssetNotFound {
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
			"error": "Failed to delete asset",
		})
	}

	return c.JSON(fiber.Map{"status": "success"})
}

// IncrementSalesCount handles internal requests to increment an asset's sales count
func (h *AssetHandler) IncrementSalesCount(c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.assetUsecase.IncrementSalesCount(c.Context(), id)
	if err != nil {
		if err == usecase.ErrAssetNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to increment sales count",
		})
	}

	return c.JSON(fiber.Map{"status": "success"})
}