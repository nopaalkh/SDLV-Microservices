package usecase

import (
	"context"
	"errors"
	"time"

	"catalog-service/internal/repository"
	"github.com/google/uuid"
)

// Custom errors
var (
	ErrAssetNotFound     = errors.New("asset not found")
	ErrInvalidAssetData  = errors.New("invalid asset data")
	ErrUnauthorized      = errors.New("unauthorized access")
)

// AssetUsecase defines the interface for asset use cases
type AssetUsecase interface {
	CreateAsset(ctx context.Context, asset *repository.Asset, creatorID string) (*repository.Asset, error)
	GetAssetByID(ctx context.Context, id string) (*repository.Asset, error)
	GetAssetsByCreator(ctx context.Context, creatorID string) ([]*repository.Asset, error)
	SearchAssets(ctx context.Context, query string, category string, sortBy string, limit int, offset int) ([]*repository.Asset, error)
	UpdateAsset(ctx context.Context, asset *repository.Asset, userID string) (*repository.Asset, error)
	DeleteAsset(ctx context.Context, id string, userID string) error
	GetAllAssets(ctx context.Context) ([]*repository.Asset, error)
	AdminDeleteAsset(ctx context.Context, id string) error
	IncrementSalesCount(ctx context.Context, id string) error
}

// assetUsecase implements AssetUsecase
type assetUsecase struct {
	assetRepo repository.AssetRepository
}

// NewAssetUsecase creates a new AssetUsecase
func NewAssetUsecase(assetRepo repository.AssetRepository) AssetUsecase {
	return &assetUsecase{assetRepo: assetRepo}
}

// CreateAsset creates a new asset
func (uc *assetUsecase) CreateAsset(ctx context.Context, asset *repository.Asset, creatorID string) (*repository.Asset, error) {
	// Validate asset data
	if asset.Title == "" || asset.Price <= 0 || asset.StoragePath == "" {
		return nil, ErrInvalidAssetData
	}

	// Generate ID
	asset.ID = uuid.New().String()
	asset.CreatedAt = time.Now()
	asset.UpdatedAt = time.Now()
	asset.CreatedBy = creatorID

	// Save asset
	err := uc.assetRepo.CreateAsset(ctx, asset)
	if err != nil {
		return nil, err
	}

	return asset, nil
}

// GetAssetByID retrieves an asset by ID
func (uc *assetUsecase) GetAssetByID(ctx context.Context, id string) (*repository.Asset, error) {
	asset, err := uc.assetRepo.FindAssetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}
	return asset, nil
}

// GetAssetsByCreator retrieves assets by creator ID
func (uc *assetUsecase) GetAssetsByCreator(ctx context.Context, creatorID string) ([]*repository.Asset, error) {
	assets, err := uc.assetRepo.FindAssetsByCreator(ctx, creatorID)
	if err != nil {
		return nil, err
	}
	return assets, nil
}

// SearchAssets searches for assets
func (uc *assetUsecase) SearchAssets(ctx context.Context, query string, category string, sortBy string, limit int, offset int) ([]*repository.Asset, error) {
	// Set default limit and offset if not provided or invalid
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	assets, err := uc.assetRepo.SearchAssets(ctx, query, category, sortBy, limit, offset)
	if err != nil {
		return nil, err
	}
	return assets, nil
}

// UpdateAsset updates an existing asset
func (uc *assetUsecase) UpdateAsset(ctx context.Context, asset *repository.Asset, userID string) (*repository.Asset, error) {
	// Find existing asset
	existingAsset, err := uc.assetRepo.FindAssetByID(ctx, asset.ID)
	if err != nil {
		return nil, err
	}
	if existingAsset == nil {
		return nil, ErrAssetNotFound
	}

	// Check ownership
	if existingAsset.CreatedBy != userID {
		return nil, ErrUnauthorized
	}

	// Validate asset data
	if asset.Title == "" || asset.Price <= 0 {
		return nil, ErrInvalidAssetData
	}

	// Update fields
	existingAsset.Title = asset.Title
	existingAsset.Description = asset.Description
	existingAsset.Price = asset.Price
	existingAsset.AIToolUsed = asset.AIToolUsed
	existingAsset.Resolution = asset.Resolution
	existingAsset.PreviewURL = asset.PreviewURL
	existingAsset.StoragePath = asset.StoragePath
	existingAsset.UpdatedAt = time.Now()

	// Save asset
	err = uc.assetRepo.UpdateAsset(ctx, existingAsset)
	if err != nil {
		return nil, err
	}

	return existingAsset, nil
}

// DeleteAsset deletes an asset
func (uc *assetUsecase) DeleteAsset(ctx context.Context, id string, userID string) error {
	// Find existing asset
	existingAsset, err := uc.assetRepo.FindAssetByID(ctx, id)
	if err != nil {
		return err
	}
	if existingAsset == nil {
		return ErrAssetNotFound
	}

	// Check ownership
	if existingAsset.CreatedBy != userID {
		return ErrUnauthorized
	}

	// Delete asset
	err = uc.assetRepo.DeleteAsset(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

// GetAllAssets retrieves all assets (admin only)
func (uc *assetUsecase) GetAllAssets(ctx context.Context) ([]*repository.Asset, error) {
	return uc.assetRepo.GetAllAssets(ctx)
}

// AdminDeleteAsset deletes an asset without ownership check (admin takedown)
func (uc *assetUsecase) AdminDeleteAsset(ctx context.Context, id string) error {
	existingAsset, err := uc.assetRepo.FindAssetByID(ctx, id)
	if err != nil {
		return err
	}
	if existingAsset == nil {
		return ErrAssetNotFound
	}
	return uc.assetRepo.DeleteAsset(ctx, id)
}

// IncrementSalesCount increments the sales count of an asset
func (uc *assetUsecase) IncrementSalesCount(ctx context.Context, id string) error {
	existingAsset, err := uc.assetRepo.FindAssetByID(ctx, id)
	if err != nil {
		return err
	}
	if existingAsset == nil {
		return ErrAssetNotFound
	}
	return uc.assetRepo.IncrementSalesCount(ctx, id)
}