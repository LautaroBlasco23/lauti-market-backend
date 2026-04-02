package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// sellerSeed holds the configuration for a store/seller account and its products.
type sellerSeed struct {
	Email       string
	Password    string
	Name        string
	Description string
	Address     string
	PhoneNumber string
	Products    []productSeed
}

// productSeed holds the configuration for a single product.
type productSeed struct {
	Name        string
	Description string
	Category    string
	Stock       int
	Price       float64
}

var sellers = []sellerSeed{
	{
		Email:       "lauti.electronics@example.com",
		Password:    "securepass1",
		Name:        "Lauti Electronics",
		Description: "Your go-to store for quality electronics and accessories.",
		Address:     "456 Tech Avenue, Buenos Aires",
		PhoneNumber: "555-0101",
		Products: []productSeed{
			{
				Name:        "Wireless Headphones",
				Description: "Over-ear noise-cancelling wireless headphones with 30-hour battery life.",
				Category:    "Audio",
				Stock:       25,
				Price:       89.99,
			},
			{
				Name:        "USB-C Hub 7-in-1",
				Description: "Compact hub with HDMI, USB-A, USB-C PD, SD card slot and more.",
				Category:    "Accessories",
				Stock:       40,
				Price:       34.99,
			},
			{
				Name:        "Mechanical Keyboard",
				Description: "Compact tenkeyless mechanical keyboard with blue switches and RGB backlight.",
				Category:    "Peripherals",
				Stock:       15,
				Price:       59.99,
			},
			{
				Name:        "LED Desk Lamp",
				Description: "Adjustable brightness LED desk lamp with USB charging port on the base.",
				Category:    "Lighting",
				Stock:       30,
				Price:       24.99,
			},
		},
	},
	{
		Email:       "green.garden.supply@example.com",
		Password:    "securepass2",
		Name:        "Green Garden Supply",
		Description: "Everything you need to grow and maintain a beautiful garden.",
		Address:     "12 Botanical Road, Mendoza",
		PhoneNumber: "555-0202",
		Products: []productSeed{
			{
				Name:        "Organic Fertilizer 5kg",
				Description: "All-natural organic fertilizer suitable for vegetables, flowers and shrubs.",
				Category:    "Fertilizers",
				Stock:       60,
				Price:       18.50,
			},
			{
				Name:        "Pruning Shears",
				Description: "Stainless steel bypass pruning shears with ergonomic grip for clean cuts.",
				Category:    "Tools",
				Stock:       45,
				Price:       14.99,
			},
			{
				Name:        "Self-Watering Planter",
				Description: "Modern self-watering planter with water reservoir for indoor plants.",
				Category:    "Planters",
				Stock:       20,
				Price:       27.00,
			},
			{
				Name:        "Heirloom Tomato Seeds",
				Description: "Pack of 50 heirloom tomato seeds, non-GMO and open-pollinated variety.",
				Category:    "Seeds",
				Stock:       100,
				Price:       4.99,
			},
			{
				Name:        "Garden Kneeling Pad",
				Description: "Thick foam kneeling pad that protects knees while gardening on hard ground.",
				Category:    "Accessories",
				Stock:       35,
				Price:       9.99,
			},
		},
	},
	{
		Email:       "urban.bookshelf@example.com",
		Password:    "securepass3",
		Name:        "Urban Bookshelf",
		Description: "A curated selection of books, stationery and reading accessories.",
		Address:     "88 Literary Lane, Cordoba",
		PhoneNumber: "555-0303",
		Products: []productSeed{
			{
				Name:        "Leather Journal A5",
				Description: "Handcrafted genuine leather journal with 200 pages of acid-free paper.",
				Category:    "Stationery",
				Stock:       50,
				Price:       22.00,
			},
			{
				Name:        "Bamboo Book Stand",
				Description: "Adjustable bamboo book stand keeps books open hands-free while reading.",
				Category:    "Accessories",
				Stock:       18,
				Price:       16.99,
			},
			{
				Name:        "Highlighter Set 8 Colors",
				Description: "Chisel tip highlighters in 8 vibrant colors, smear-proof and quick-drying.",
				Category:    "Stationery",
				Stock:       75,
				Price:       8.49,
			},
		},
	},
}

func main() {
	_ = godotenv.Load()

	baseURL := flag.String("base-url", "http://localhost:8000", "Base URL of the running backend")
	flag.Parse()

	unsplashKey := os.Getenv("UNSPLASH_ACCESS_KEY")
	if unsplashKey == "" {
		log.Println("UNSPLASH_ACCESS_KEY not set; products will be created without images")
	}

	client := &http.Client{Timeout: 30 * time.Second}

	log.Printf("Checking server health at %s/health ...", *baseURL)
	resp, err := client.Get(*baseURL + "/health")
	if err != nil {
		log.Fatalf("Server is not reachable at %s: %v", *baseURL, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Health check returned unexpected status %d", resp.StatusCode)
	}
	log.Println("Server is reachable. Starting seed...")

	suffix := fmt.Sprintf("%05d", rand.Intn(100000))
	log.Printf("Run suffix: %s", suffix)

	var totalStores, totalProducts, failedStores, failedProducts int

	for _, seller := range sellers {
		s := seller
		// inject suffix so each run produces unique emails and names
		s.Email = strings.Replace(s.Email, "@", "+"+suffix+"@", 1)
		s.Name = s.Name + " " + suffix

		accountID, ok := registerStore(client, *baseURL, s)
		if !ok {
			failedStores++
			continue
		}
		totalStores++

		for _, product := range s.Products {
			p := product
			p.Name = p.Name + " " + suffix
			if createProduct(client, *baseURL, accountID, p, unsplashKey) {
				totalProducts++
			} else {
				failedProducts++
			}
		}
	}

	fmt.Println()
	log.Printf("Seed complete: %d stores created (%d failed), %d products created (%d failed)",
		totalStores, failedStores, totalProducts, failedProducts)
}

// registerStore POSTs to /auth/register/store and returns the account_id on success.
func registerStore(client *http.Client, baseURL string, seller sellerSeed) (string, bool) {
	payload := map[string]string{
		"email":        seller.Email,
		"password":     seller.Password,
		"name":         seller.Name,
		"description":  seller.Description,
		"address":      seller.Address,
		"phone_number": seller.PhoneNumber,
	}

	body, status, err := postJSON(client, baseURL+"/auth/register/store", payload)
	if err != nil {
		log.Printf("[store] ERROR registering %q: %v", seller.Name, err)
		return "", false
	}

	switch status {
	case http.StatusCreated:
		accountID, _ := body["account_id"].(string)
		log.Printf("[store] Created %q (account_id=%s)", seller.Name, accountID)
		return accountID, true
	case http.StatusConflict:
		log.Printf("[store] SKIP %q — email %q already exists (409)", seller.Name, seller.Email)
		return "", false
	default:
		log.Printf("[store] ERROR registering %q — status %d, body: %v", seller.Name, status, body)
		return "", false
	}
}

// createProduct POSTs multipart form data to /stores/{storeID}/products, then
// patches the image_url via PUT if an Unsplash photo is found.
func createProduct(client *http.Client, baseURL, storeID string, product productSeed, unsplashKey string) bool {
	fields := map[string]string{
		"name":        product.Name,
		"description": product.Description,
		"category":    product.Category,
		"stock":       strconv.Itoa(product.Stock),
		"price":       strconv.FormatFloat(product.Price, 'f', 2, 64),
	}

	endpoint := fmt.Sprintf("%s/stores/%s/products", baseURL, storeID)
	body, status, err := postMultipart(client, endpoint, fields)
	if err != nil {
		log.Printf("[product] ERROR creating %q: %v", product.Name, err)
		return false
	}
	if status != http.StatusCreated {
		log.Printf("[product] ERROR creating %q — status %d, body: %v", product.Name, status, body)
		return false
	}

	productID, _ := body["id"].(string)
	log.Printf("[product] Created %q (id=%s)", product.Name, productID)

	if unsplashKey == "" || productID == "" {
		return true
	}

	imageURL, err := fetchUnsplashURL(client, unsplashKey, product.Name)
	if err != nil {
		log.Printf("[product] WARNING: failed to fetch image for %q: %v", product.Name, err)
		return true
	}
	if imageURL == "" {
		log.Printf("[product] WARNING: no image found on Unsplash for %q", product.Name)
		return true
	}

	putEndpoint := fmt.Sprintf("%s/stores/%s/products/%s", baseURL, storeID, productID)
	putPayload := map[string]any{
		"name":        product.Name,
		"description": product.Description,
		"category":    product.Category,
		"stock":       product.Stock,
		"price":       product.Price,
		"image_url":   imageURL,
	}
	_, putStatus, err := putJSON(client, putEndpoint, putPayload)
	if err != nil || putStatus != http.StatusOK {
		log.Printf("[product] WARNING: failed to set image_url for %q (status %d): %v", product.Name, putStatus, err)
		return true
	}
	log.Printf("[product] Image set for %q", product.Name)
	return true
}

// fetchUnsplashURL queries the Unsplash Search Photos API and returns the URL
// of the first result's regular-sized image. Returns "" when no results found.
func fetchUnsplashURL(client *http.Client, accessKey, query string) (string, error) {
	apiURL := "https://api.unsplash.com/search/photos?query=" + url.QueryEscape(query) + "&per_page=1"

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating unsplash request: %w", err)
	}
	req.Header.Set("Authorization", "Client-ID "+accessKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling unsplash API: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Results []struct {
			URLs struct {
				Regular string `json:"regular"`
			} `json:"urls"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("decoding unsplash response: %w", err)
	}

	if len(apiResp.Results) == 0 {
		return "", nil
	}
	return apiResp.Results[0].URLs.Regular, nil
}

// postMultipart sends a multipart/form-data POST request with the given fields.
func postMultipart(client *http.Client, rawURL string, fields map[string]string) (map[string]any, int, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, 0, fmt.Errorf("writing field %q: %w", key, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, 0, fmt.Errorf("closing multipart writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, rawURL, &buf)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return doRequest(client, req)
}

// putJSON marshals payload as JSON and sends a PUT request.
func putJSON(client *http.Client, rawURL string, payload any) (map[string]any, int, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("marshalling payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, rawURL, bytes.NewReader(data))
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return doRequest(client, req)
}

// postJSON marshals payload as JSON and sends a POST request. It returns the
// decoded response body, HTTP status code, and any transport-level error.
func postJSON(client *http.Client, url string, payload any) (map[string]any, int, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("marshalling payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return doRequest(client, req)
}

// doRequest executes the HTTP request and decodes the JSON response body.
func doRequest(client *http.Client, req *http.Request) (map[string]any, int, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response body: %w", err)
	}

	var result map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &result); err != nil {
			// Return the raw text in a synthetic map so callers can log it.
			result = map[string]any{"raw": string(raw)}
		}
	}

	return result, resp.StatusCode, nil
}
