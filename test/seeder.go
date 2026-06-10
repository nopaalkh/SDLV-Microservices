//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var gatewayURL = "http://localhost:8080"

func main() {
	client := &http.Client{Timeout: 10 * time.Second}
	fmt.Println("=== Starting Database Seeder ===")

	// 1. Seed Admin (from .env: admin@marketplace.com / admin123456)
	// This user is auto-created by identity-service on startup, so just login
	adminToken := loginUser(client, "admin@marketplace.com", "admin123456")
	if adminToken == "" {
		log.Println("WARN: Could not login as admin. Admin may not be seeded yet.")
	} else {
		fmt.Println("✓ Admin login successful")
	}

	// 2. Seed Creators
	fmt.Println("\nSeeding Creators...")
	seedUser(client, "nox@studio.com", "password123", "CREATOR")
	seedUser(client, "audiohz@music.com", "password123", "CREATOR")
	seedUser(client, "polygen@3d.com", "password123", "CREATOR")

	creatorNoxToken := loginUser(client, "nox@studio.com", "password123")
	creatorAudioHzToken := loginUser(client, "audiohz@music.com", "password123")
	creatorPolygenToken := loginUser(client, "polygen@3d.com", "password123")

	// 3. Seed Buyers
	fmt.Println("\nSeeding Buyers...")
	seedUser(client, "buyer@example.com", "password123", "BUYER")

	// 4. Seed Assets with Rupiah prices
	fmt.Println("\nSeeding Assets (Rupiah prices)...")
	assets := []map[string]interface{}{
		{
			"token":       creatorNoxToken,
			"title":       "Cyberpunk UI Kit Premium",
			"description": "UI Kit futuristik lengkap dengan lebih dari 200 komponen, efek neon, dan tata letak mode gelap. Cocok untuk web dan aplikasi mobile.",
			"price":       299000,
			"category":    "UI Kit",
			"preview_url": "https://images.unsplash.com/photo-1618761714954-0b8cd0026356?auto=format&fit=crop&w=1200&q=80",
			"storage_path": "mock/cyberpunk.zip",
		},
		{
			"token":       creatorAudioHzToken,
			"title":       "Synthwave Audio Pack",
			"description": "Koleksi musik loop retro 80-an, one-shot, dan trek ambient untuk game dan video Anda. Tersedia 50+ track dalam format WAV berkualitas tinggi.",
			"price":       149000,
			"category":    "Audio",
			"preview_url": "https://images.unsplash.com/photo-1511379938547-c1f69419868d?auto=format&fit=crop&w=500&q=80",
			"storage_path": "mock/synthwave.zip",
		},
		{
			"token":       creatorPolygenToken,
			"title":       "3D Abstract Shapes",
			"description": "Bentuk 3D abstrak resolusi tinggi untuk branding modern, poster, dan latar belakang web. Termasuk file Blender dan resolusi 4K.",
			"price":       459000,
			"category":    "3D Model",
			"preview_url": "https://images.unsplash.com/photo-1550751827-4bd374c3f58b?auto=format&fit=crop&w=500&q=80",
			"storage_path": "mock/abstract.zip",
		},
		{
			"token":       creatorNoxToken,
			"title":       "Minimalist Icon Set",
			"description": "Set ikon minimalis dengan 500+ ikon dalam format SVG. Sempurna untuk dashboard, aplikasi mobile, dan desain web modern.",
			"price":       199000,
			"category":    "Icon Pack",
			"preview_url": "https://images.unsplash.com/photo-1558655146-9f40138edfeb?auto=format&fit=crop&w=500&q=80",
			"storage_path": "mock/icons.zip",
		},
		{
			"token":       creatorAudioHzToken,
			"title":       "Lo-Fi Chill Beats",
			"description": "Koleksi musik Lo-Fi chill untuk konten kreator. 30 trek full-length bebas royalti, cocok untuk YouTube dan podcast.",
			"price":       89000,
			"category":    "Audio",
			"preview_url": "https://images.unsplash.com/photo-1493225457124-a3eb161ffa5f?auto=format&fit=crop&w=500&q=80",
			"storage_path": "mock/lofi.zip",
		},
		{
			"token":       creatorPolygenToken,
			"title":       "Neon Gradient Textures",
			"description": "Tekstur gradien neon beresolusi tinggi. 40+ file PNG/JPEG siap pakai untuk desain grafis, social media, dan web.",
			"price":       79000,
			"category":    "Textures",
			"preview_url": "https://images.unsplash.com/photo-1557672172-298e090bd0f1?auto=format&fit=crop&w=500&q=80",
			"storage_path": "mock/textures.zip",
		},
	}

	for _, a := range assets {
		token := a["token"].(string)
		delete(a, "token")
		if token != "" {
			createAsset(client, token, a)
		} else {
			fmt.Printf("  ✗ Skipped (no token): %s\n", a["title"])
		}
	}

	fmt.Println("\n=== Seeding Completed Successfully! ===")
	fmt.Println("\nAkun yang tersedia:")
	fmt.Println("  Admin:   admin@marketplace.com / admin123456")
	fmt.Println("  Kreator: nox@studio.com / password123")
	fmt.Println("  Kreator: audiohz@music.com / password123")
	fmt.Println("  Kreator: polygen@3d.com / password123")
	fmt.Println("  Pembeli: buyer@example.com / password123")
}

func seedUser(client *http.Client, email, password, role string) {
	registerData := map[string]string{
		"email":    email,
		"password": password,
		"role":     role,
	}
	jsonData, _ := json.Marshal(registerData)
	resp, err := client.Post(gatewayURL+"/api/auth/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("  ✗ Register %s: %v\n", email, err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("  ✓ Registered: %s (%s)\n", email, role)
	} else if resp.StatusCode == http.StatusConflict {
		fmt.Printf("  ~ Already exists: %s\n", email)
	} else {
		fmt.Printf("  ✗ Register %s: status %d\n", email, resp.StatusCode)
	}
}

func loginUser(client *http.Client, email, password string) string {
	loginData := map[string]string{
		"email":    email,
		"password": password,
	}
	jsonData, _ := json.Marshal(loginData)
	resp, err := client.Post(gatewayURL+"/api/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("  ✗ Login %s: %v", email, err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("  ✗ Login %s: status %d\n", email, resp.StatusCode)
		return ""
	}

	var authResp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&authResp)
	fmt.Printf("  ✓ Logged in: %s\n", email)
	return authResp.Token
}

func createAsset(client *http.Client, token string, data map[string]interface{}) {
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", gatewayURL+"/api/assets", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  ✗ Asset '%s': %v\n", data["title"], err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("  ✓ Created: %s (Rp %v)\n", data["title"], data["price"])
	} else {
		fmt.Printf("  ✗ Asset '%s': status %d\n", data["title"], resp.StatusCode)
	}
}
