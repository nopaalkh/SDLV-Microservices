package usecase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OrderClient interface {
	HasPurchased(userID, assetID string) (bool, error)
}

type orderClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewOrderClient(baseURL string, apiKey string) OrderClient {
	return &orderClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *orderClient) HasPurchased(userID, assetID string) (bool, error) {
	url := fmt.Sprintf("%s/api/orders/internal/check-purchase?user_id=%s&asset_id=%s", c.baseURL, userID, assetID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	
	req.Header.Set("X-Internal-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		HasPurchased bool `json:"has_purchased"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.HasPurchased, nil
}
