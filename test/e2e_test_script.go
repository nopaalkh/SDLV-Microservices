//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"os"
)

var (
	gatewayURL       = getEnvOrDefault("GATEWAY_URL", "http://localhost:8080")
	identityService  = gatewayURL
	catalogService   = gatewayURL
	orderService     = gatewayURL
	testUserPassword = "password123"
)

func getEnvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var testUserEmail = fmt.Sprintf("test%d@example.com", time.Now().Unix())

type AuthResponse struct {
	Token string `json:"token"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type AssetResponse struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type OrderResponse struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	AssetID    string  `json:"asset_id"`
	Quantity   int     `json:"quantity"`
	TotalPrice float64 `json:"total_price"`
	Status     string  `json:"status"`
}

func main() {
	client := &http.Client{Timeout: 10 * time.Second}

	// 1. Register and login user
	fmt.Println("=== Testing User Registration and Login ===")
	userID, token := registerAndLoginUser(client)
	fmt.Printf("User registered with ID: %s\n", userID)
	fmt.Printf("Authentication token: %s\n", token)

	// 2. Create an asset
	fmt.Println("\n=== Testing Asset Creation ===")
	createAsset(client, token)

	// 3. Get available assets
	fmt.Println("\n=== Testing Asset Search ===")
	assetID := getAvailableAssets(client, token)
	fmt.Printf("Selected asset ID: %s\n", assetID)

	// 4. Create order
	fmt.Println("\n=== Testing Order Creation ===")
	orderID := createOrder(client, token, assetID)
	fmt.Printf("Order created with ID: %s\n", orderID)

	// 4. Test payment flow
	fmt.Println("\n=== Testing Payment Flow ===")
	testPaymentFlow(client, token, orderID)

	// 5. Test Admin flow
	fmt.Println("\n=== Testing Admin Flow ===")
	adminToken := testAdminFlow(client)

	// 6. Test Review flow
	fmt.Println("\n=== Testing Review Flow ===")
	testReviewFlow(client, token, assetID)

	// 7. Test Withdrawal flow
	fmt.Println("\n=== Testing Withdrawal Flow ===")
	testWithdrawalFlow(client, token, adminToken)

	fmt.Println("\n=== End-to-End Test Completed Successfully! ===")
}

func registerAndLoginUser(client *http.Client) (string, string) {
	// Register user
	registerData := map[string]string{
		"email":    testUserEmail,
		"password": testUserPassword,
		"role":     "CREATOR",
	}

	jsonData, _ := json.Marshal(registerData)
	resp, err := client.Post(identityService+"/api/auth/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to register user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to register user: %s, %s", resp.Status, string(body))
	}

	// Login user
	loginData := map[string]string{
		"email":    testUserEmail,
		"password": testUserPassword,
	}

	jsonData, _ = json.Marshal(loginData)
	resp, err = client.Post(identityService+"/api/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to login: %s, %s", resp.Status, string(body))
	}

	var authResp AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		log.Fatalf("Failed to decode auth response: %v", err)
	}

	// Get user details
	req, err := http.NewRequest("GET", identityService+"/api/auth/me", nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+authResp.Token)

	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("Failed to get user details: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to get user details: %s, %s", resp.Status, string(body))
	}

	var userResp UserResponse
	err = json.NewDecoder(resp.Body).Decode(&userResp)
	if err != nil {
		log.Fatalf("Failed to decode user response: %v", err)
	}

	return userResp.ID, authResp.Token
}

func createAsset(client *http.Client, token string) {
	assetData := map[string]interface{}{
		"title":       "Test Asset",
		"description": "A wonderful test asset",
		"price":       100.50,
		"storage_path":    "https://example.com/asset.png",
	}

	jsonData, _ := json.Marshal(assetData)
	req, err := http.NewRequest("POST", catalogService+"/api/assets", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to create asset: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to create asset: %s, %s", resp.Status, string(body))
	}
	fmt.Println("Asset created successfully")
}

func getAvailableAssets(client *http.Client, token string) string {
	searchData := map[string]string{"query": ""}
	jsonData, _ := json.Marshal(searchData)
	req, err := http.NewRequest("POST", catalogService+"/api/assets/search", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to get assets: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to get assets: %s, %s", resp.Status, string(body))
	}

	var respObj struct {
		Assets []AssetResponse `json:"assets"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respObj)
	if err != nil {
		log.Fatalf("Failed to decode assets response: %v", err)
	}

	if len(respObj.Assets) == 0 {
		log.Fatal("No assets available")
	}

	return respObj.Assets[0].ID
}

func createOrder(client *http.Client, token string, assetID string) string {
	orderData := map[string]interface{}{
		"asset_id":  assetID,
		"quantity":  1,
	}

	jsonData, _ := json.Marshal(orderData)
	req, err := http.NewRequest("POST", orderService+"/api/orders", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to create order: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to create order: %s, %s", resp.Status, string(body))
	}

	var orderResp OrderResponse
	err = json.NewDecoder(resp.Body).Decode(&orderResp)
	if err != nil {
		log.Fatalf("Failed to decode order response: %v", err)
	}

	return orderResp.ID
}

func testPaymentFlow(client *http.Client, token string, orderID string) {
	// Test process payment
	req, err := http.NewRequest("PUT", orderService+"/api/orders/"+orderID+"/mock-pay", nil)
	if err != nil {
		log.Fatalf("Failed to create payment request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to process payment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to process payment: %s, %s", resp.Status, string(body))
	}

	var paymentResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&paymentResp)
	if err != nil {
		log.Fatalf("Failed to decode payment response: %v", err)
	}

	fmt.Printf("Payment processed successfully. Response: %+v\n", paymentResp)

	// Verify order status
	req, err = http.NewRequest("GET", orderService+"/api/orders/"+orderID, nil)
	if err != nil {
		log.Fatalf("Failed to create order status request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("Failed to get order status: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to get order status: %s, %s", resp.Status, string(body))
	}

	var orderResp OrderResponse
	err = json.NewDecoder(resp.Body).Decode(&orderResp)
	if err != nil {
		log.Fatalf("Failed to decode order response: %v", err)
	}

	fmt.Printf("Final order status: %s\n", orderResp.Status)
}

func testAdminFlow(client *http.Client) string {
	// Login as admin
	loginData := map[string]string{
		"email":    "admin@marketplace.com",
		"password": "admin123456",
	}

	jsonData, _ := json.Marshal(loginData)
	req, err := http.NewRequest("POST", identityService+"/api/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to create admin login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Admin login failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Admin login failed: %s, %s", resp.Status, string(body))
	}

	var authResp AuthResponse
	json.NewDecoder(resp.Body).Decode(&authResp)
	adminToken := authResp.Token
	fmt.Println("Admin logged in successfully")

	// Get all users
	req, _ = http.NewRequest("GET", identityService+"/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatalf("Admin get users failed: %v", err)
	}
	fmt.Println("Admin retrieved users successfully")

	// Get all assets
	req, _ = http.NewRequest("GET", catalogService+"/api/admin/assets", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatalf("Admin get assets failed: %v", err)
	}
	fmt.Println("Admin retrieved assets successfully")

	// Get all orders
	req, _ = http.NewRequest("GET", orderService+"/api/admin/orders", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatalf("Admin get orders failed: %v", err)
	}
	fmt.Println("Admin retrieved orders successfully")

	// Get platform revenue
	req, _ = http.NewRequest("GET", orderService+"/api/admin/orders/revenue", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Admin get platform revenue failed: %s, %s", resp.Status, string(body))
	}
	var revResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&revResp)
	fmt.Printf("Admin retrieved platform revenue successfully: %.2f %s\n", revResp["platform_revenue"], revResp["currency"])


	return adminToken
}

func testReviewFlow(client *http.Client, token string, assetID string) {
	reviewData := map[string]interface{}{
		"rating":  5,
		"comment": "This is a fantastic asset, highly recommended!",
	}
	jsonData, _ := json.Marshal(reviewData)
	req, _ := http.NewRequest("POST", catalogService+"/api/assets/"+assetID+"/reviews", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to create review: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to create review: %s, %s", resp.Status, string(body))
	}
	fmt.Println("Review created successfully")
}

func testWithdrawalFlow(client *http.Client, token string, adminToken string) {
	// Creator requests withdrawal
	withdrawalData := map[string]interface{}{
		"amount":         10.0,
		"bank_code":      "BCA",
		"account_number": "1234567890",
	}
	jsonData, _ := json.Marshal(withdrawalData)
	req, _ := http.NewRequest("POST", orderService+"/api/withdrawals", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to request withdrawal: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to request withdrawal: %s, %s", resp.Status, string(body))
	}
	
	var withdrawal map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&withdrawal)
	withdrawalID := withdrawal["id"].(string)
	fmt.Printf("Withdrawal requested successfully with ID: %s\n", withdrawalID)

	// Admin approves withdrawal
	req, _ = http.NewRequest("PUT", orderService+"/api/admin/withdrawals/"+withdrawalID+"/approve", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("Failed to approve withdrawal: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to approve withdrawal: %s, %s", resp.Status, string(body))
	}
	fmt.Println("Withdrawal approved by Admin successfully")
}