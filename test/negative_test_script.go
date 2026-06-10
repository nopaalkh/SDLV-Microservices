//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	identityURL = "http://localhost:3001/api"
	catalogURL  = "http://localhost:3002/api"
	orderURL    = "http://localhost:3003/api"
	adminToken  string
	buyerToken  string
	buyerID     string
)

func main() {
	fmt.Println("=== Starting Negative Testing ===")

	client := &http.Client{Timeout: 10 * time.Second}

	// Setup: Need an admin token and a buyer token
	adminToken = login(client, identityURL+"/auth/login", "admin@marketplace.com", "admin123456")
	if adminToken == "" {
		fmt.Println("❌ Failed to login as admin. Check credentials.")
		return
	}

	buyerToken, buyerID = registerAndLogin(client, identityURL+"/auth/register", identityURL+"/auth/login", fmt.Sprintf("buyer_%d@test.com", time.Now().Unix()), "password123", "BUYER")
	if buyerToken == "" {
		fmt.Println("❌ Failed to login as buyer.")
		return
	}

	// Test 1: Unauthorized Access (No Token)
	fmt.Println("\n--- Test 1: Unauthorized Access (No Token) ---")
	status, _ := sendRequest(client, "GET", orderURL+"/orders/my-orders", nil, "")
	assertStatus("Unauthorized Access", status, http.StatusUnauthorized)

	// Test 2: Invalid JWT Token
	fmt.Println("\n--- Test 2: Invalid JWT Token ---")
	status, _ = sendRequest(client, "GET", orderURL+"/orders/my-orders", nil, "Bearer invalid.token.here")
	assertStatus("Invalid Token", status, http.StatusUnauthorized)

	// Test 3: Protected Internal Endpoint (No API Key)
	fmt.Println("\n--- Test 3: Protected Internal Endpoint (No API Key) ---")
	status, _ = sendRequest(client, "GET", identityURL+"/users/"+buyerID, nil, "")
	assertStatus("Internal API Protection", status, http.StatusUnauthorized)

	// Test 4: Role-based Access Control (Buyer accessing Admin endpoint)
	fmt.Println("\n--- Test 4: RBAC Violation (Buyer accessing Admin) ---")
	status, _ = sendRequest(client, "GET", orderURL+"/admin/orders/revenue", nil, "Bearer "+buyerToken)
	assertStatus("Admin Route Protection", status, http.StatusForbidden)

	// Test 5: Order Validation (Asset does not exist)
	fmt.Println("\n--- Test 5: Order Validation (Invalid Asset) ---")
	orderPayload := map[string]interface{}{
		"asset_id": "invalid-asset-id-123",
		"quantity": 1,
	}
	status, body := sendRequest(client, "POST", orderURL+"/orders", orderPayload, "Bearer "+buyerToken)
	assertStatus("Order Invalid Asset", status, http.StatusNotFound)
	fmt.Printf("Error response: %s\n", body)

	// Test 6: Unique Review Constraint
	fmt.Println("\n--- Test 6: Unique Review Constraint ---")
	// 6.1 Create an asset first (Need CREATOR role)
	creatorToken, _ := registerAndLogin(client, identityURL+"/auth/register", identityURL+"/auth/login", fmt.Sprintf("creator_%d@test.com", time.Now().Unix()), "password123", "CREATOR")
	
	assetPayload := map[string]interface{}{
		"title": "Test Asset", "description": "Desc", "price": 1000, "storage_path": "/test/path", "category": "Art",
	}
	_, assetBody := sendRequest(client, "POST", catalogURL+"/assets", assetPayload, "Bearer "+creatorToken)
	var assetResp map[string]interface{}
	json.Unmarshal([]byte(assetBody), &assetResp)
	assetID := assetResp["id"].(string)

	// 6.2 Buyer buys the asset to be able to review
	orderPayload = map[string]interface{}{"asset_id": assetID, "quantity": 1}
	_, orderBody := sendRequest(client, "POST", orderURL+"/orders", orderPayload, "Bearer "+buyerToken)
	var orderResp map[string]interface{}
	json.Unmarshal([]byte(orderBody), &orderResp)
	orderID := orderResp["id"].(string)

	// Mock pay to make order PAID
	sendRequest(client, "PUT", orderURL+"/orders/"+orderID+"/mock-pay", nil, "Bearer "+buyerToken)

	// 6.3 Buyer submits first review
	reviewPayload := map[string]interface{}{"rating": 5, "comment": "Great!"}
	status1, _ := sendRequest(client, "POST", catalogURL+"/assets/"+assetID+"/reviews", reviewPayload, "Bearer "+buyerToken)
	if status1 == 201 {
		fmt.Println("✅ First review created successfully")
	}

	// 6.4 Buyer submits second review (Should Fail)
	status2, body2 := sendRequest(client, "POST", catalogURL+"/assets/"+assetID+"/reviews", reviewPayload, "Bearer "+buyerToken)
	assertStatus("Duplicate Review Protection", status2, http.StatusInternalServerError) // Because it returns DB error 500 when constraint fails, or 400 if mapped
	fmt.Printf("Error response on duplicate review: %s\n", body2)

	// Test 7: Withdrawal Validation (Insufficient Balance)
	fmt.Println("\n--- Test 7: Withdrawal Validation (Insufficient Balance) ---")
	withdrawPayload := map[string]interface{}{
		"amount": 999999999, // Too high
		"bank_code": "BCA",
		"account_number": "123456",
	}
	status, body = sendRequest(client, "POST", orderURL+"/withdrawals", withdrawPayload, "Bearer "+creatorToken)
	assertStatus("Insufficient Balance", status, http.StatusBadRequest)
	fmt.Printf("Error response: %s\n", body)

	fmt.Println("\n=== Negative Testing Completed ===")
}

func assertStatus(testName string, actual, expected int) {
	if actual == expected || (expected == http.StatusInternalServerError && actual >= 400) { // Tolerate 500 or 400 for DB errors
		fmt.Printf("✅ %s: PASS (Got %d)\n", testName, actual)
	} else {
		fmt.Printf("❌ %s: FAIL (Expected %d, Got %d)\n", testName, expected, actual)
	}
}

func sendRequest(client *http.Client, method, url string, payload interface{}, authHeader string) (int, string) {
	var bodyReader io.Reader
	if payload != nil {
		jsonPayload, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(jsonPayload)
	}

	req, _ := http.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(bodyBytes)
}

func login(client *http.Client, url, email, password string) string {
	payload := map[string]string{"email": email, "password": password}
	_, body := sendRequest(client, "POST", url, payload, "")
	var resp map[string]interface{}
	json.Unmarshal([]byte(body), &resp)
	if token, ok := resp["token"].(string); ok {
		return token
	}
	return ""
}

func registerAndLogin(client *http.Client, regUrl, loginUrl, email, password, role string) (string, string) {
	payload := map[string]string{"email": email, "password": password, "role": role}
	_, regBody := sendRequest(client, "POST", regUrl, payload, "")
	var regResp map[string]interface{}
	json.Unmarshal([]byte(regBody), &regResp)
	
	id := ""
	if idVal, ok := regResp["id"].(string); ok {
		id = idVal
	} else if user, ok := regResp["user"].(map[string]interface{}); ok {
		id = user["id"].(string)
	}

	return login(client, loginUrl, email, password), id
}
