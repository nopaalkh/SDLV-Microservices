package usecase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type IdentityClient interface {
	GetUserEmail(userID string) (string, error)
}

type identityClient struct {
	identityServiceURL string
	internalAPIKey     string
	httpClient         *http.Client
}

func NewIdentityClient(identityServiceURL, internalAPIKey string) IdentityClient {
	return &identityClient{
		identityServiceURL: identityServiceURL,
		internalAPIKey:     internalAPIKey,
		httpClient:         &http.Client{Timeout: 5 * time.Second},
	}
}

type UserResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (c *identityClient) GetUserEmail(userID string) (string, error) {
	url := fmt.Sprintf("%s/api/users/%s", c.identityServiceURL, userID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Internal-API-Key", c.internalAPIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get user: status %d", resp.StatusCode)
	}

	var userResp UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return "", err
	}

	if userResp.Name != "" {
		return userResp.Name, nil
	}
	return userResp.Email, nil
}
