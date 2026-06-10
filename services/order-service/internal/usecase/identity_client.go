package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// IdentityClient handles communication with the Identity Service
type IdentityClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewIdentityClient creates a new IdentityClient
func NewIdentityClient(baseURL string, apiKey string) *IdentityClient {
	return &IdentityClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// GetUserEmail gets the user email from Identity Service
func (ic *IdentityClient) GetUserEmail(ctx context.Context, userID string) (string, error) {
	url := fmt.Sprintf("%s/api/users/%s", ic.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Internal-API-Key", ic.apiKey)

	resp, err := ic.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call identity service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("identity service returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var user struct {
		Email string `json:"email"`
	}
	err = json.Unmarshal(body, &user)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return user.Email, nil
}
