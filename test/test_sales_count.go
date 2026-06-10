package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const gatewayURL = "http://localhost:8080"

type RegisterReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRes struct {
	Token string `json:"token"`
}

type AssetReq struct {
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	StoragePath string  `json:"storage_path"`
}

type AssetRes struct {
	ID         string  `json:"id"`
	SalesCount int     `json:"sales_count"`
	Price      float64 `json:"price"`
}

type OrderReq struct {
	AssetID  string `json:"asset_id"`
	Quantity int    `json:"quantity"`
}

type OrderRes struct {
	ID string `json:"id"`
}

func main() {
	timestamp := time.Now().Unix()
	creatorEmail := fmt.Sprintf("creator%d@test.com", timestamp)
	buyerEmail := fmt.Sprintf("buyer%d@test.com", timestamp)
	password := "password123"

	// Wait for services to be ready
	time.Sleep(3 * time.Second)

	fmt.Println("1. Register Creator")
	doReq("POST", "/api/auth/register", "", RegisterReq{Email: creatorEmail, Password: password, Role: "CREATOR"}, nil)

	fmt.Println("2. Login Creator")
	var loginRes LoginRes
	doReq("POST", "/api/auth/login", "", LoginReq{Email: creatorEmail, Password: password}, &loginRes)
	creatorToken := loginRes.Token
	fmt.Printf("Creator Token length: %d\n", len(creatorToken))

	fmt.Println("3. Create Asset")
	var assetRes AssetRes
	doReq("POST", "/api/assets/", creatorToken, AssetReq{Title: "Test Asset", Price: 100.0, StoragePath: "/tmp/test.zip"}, &assetRes)
	assetID := assetRes.ID
	fmt.Printf("Created Asset ID: %s, Initial SalesCount: %d\n", assetID, assetRes.SalesCount)

	fmt.Println("4. Register Buyer")
	doReq("POST", "/api/auth/register", "", RegisterReq{Email: buyerEmail, Password: password, Role: "BUYER"}, nil)

	fmt.Println("5. Login Buyer")
	doReq("POST", "/api/auth/login", "", LoginReq{Email: buyerEmail, Password: password}, &loginRes)
	buyerToken := loginRes.Token
	fmt.Printf("Buyer Token length: %d\n", len(buyerToken))

	fmt.Println("6. Create Order")
	var orderRes OrderRes
	doReq("POST", "/api/orders/", buyerToken, OrderReq{AssetID: assetID, Quantity: 1}, &orderRes)
	orderID := orderRes.ID
	fmt.Printf("Created Order ID: %s\n", orderID)

	fmt.Println("7. Mock Payment (Wait a moment before processing)")
	time.Sleep(1 * time.Second)
	doReq("PUT", fmt.Sprintf("/api/orders/%s/mock-pay", orderID), buyerToken, nil, nil)

	fmt.Println("8. Wait for asynchronous update (Wait 2 seconds)")
	time.Sleep(2 * time.Second)

	fmt.Println("9. Fetch Asset and Check Sales Count")
	var finalAsset AssetRes
	doReq("GET", fmt.Sprintf("/api/assets/%s", assetID), "", nil, &finalAsset)
	
	fmt.Printf("Final Asset ID: %s\n", finalAsset.ID)
	fmt.Printf("Final Sales Count: %d\n", finalAsset.SalesCount)

	if finalAsset.SalesCount == 1 {
		fmt.Println("\n✅ TEST PASSED: Sales count successfully incremented after payment!")
	} else {
		fmt.Println("\n❌ TEST FAILED: Sales count was not incremented correctly.")
	}
}

func doReq(method, path, token string, body interface{}, out interface{}) {
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, gatewayURL+path, bodyReader)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fmt.Printf("Error %d: %s\n", resp.StatusCode, string(respBody))
	} else if out != nil {
		json.Unmarshal(respBody, out)
	}
}
