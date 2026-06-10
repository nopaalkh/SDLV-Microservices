package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AssetResponse represents the response from Catalog Service
type AssetResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	CreatedBy string  `json:"created_by"`
}

// assetService implements AssetService interface
type assetService struct {
	catalogServiceURL string
	internalAPIKey    string
	client            *http.Client
}

// NewAssetService creates a new AssetService
func NewAssetService(catalogServiceURL string, internalAPIKey string) AssetService {
	return &assetService{
		catalogServiceURL: catalogServiceURL,
		internalAPIKey:    internalAPIKey,
		client:            &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAssetDetails gets the details of an asset including price and creator from the Catalog Service
func (s *assetService) GetAssetDetails(ctx context.Context, assetID string) (*AssetResponse, error) {
	url := fmt.Sprintf("%s/api/assets/%s", s.catalogServiceURL, assetID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call catalog service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("asset not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog service returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var assetResp AssetResponse
	if err := json.Unmarshal(body, &assetResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &assetResp, nil
}

// ValidateAsset validates if an asset exists in the Catalog Service
func (s *assetService) ValidateAsset(ctx context.Context, assetID string) (bool, error) {
	url := fmt.Sprintf("%s/api/assets/%s", s.catalogServiceURL, assetID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to call catalog service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("catalog service returned status: %d", resp.StatusCode)
	}

	return true, nil
}

// IncrementAssetSales increments the sales count of an asset
func (s *assetService) IncrementAssetSales(ctx context.Context, assetID string) error {
	url := fmt.Sprintf("%s/api/assets/internal/%s/increment-sales", s.catalogServiceURL, assetID)

	req, err := http.NewRequestWithContext(ctx, "PUT", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Internal-API-Key", s.internalAPIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call catalog service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("catalog service returned status: %d", resp.StatusCode)
	}

	return nil
}
