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

// sellerSeed holds the data needed to register a store and seed its catalog.
type sellerSeed struct {
	Email       string
	Password    string
	Name        string
	Description string
	Address     string
	PhoneNumber string
	Products    []productSeed
}

// productSeed holds a generated product instance ready to be POSTed.
type productSeed struct {
	Name        string
	Description string
	Category    string
	Stock       int
	Price       float64
	ImageQuery  string // search term used against Unsplash for the product image
}

// productTemplate is the blueprint a product instance is generated from.
type productTemplate struct {
	BaseName    string
	ImageQuery  string
	Description string
	MinPrice    float64
	MaxPrice    float64
}

// categoryDef groups everything needed to spin up a randomized seller of a given category.
type categoryDef struct {
	Name              string
	StoreNames        []string
	StoreDescriptions []string
	Products          []productTemplate
}

var productAdjectives = []string{
	"Premium", "Classic", "Modern", "Compact", "Pro", "Eco",
	"Vintage", "Deluxe", "Essential", "Urban", "Smart", "Everyday",
	"Heritage", "Signature", "Travel",
}

var modelCodes = []string{"X1", "X2", "Mk II", "Pro", "v2", "Plus", "Lite", "S", "Edge", "Air"}

var cities = []string{
	"Buenos Aires", "Cordoba", "Mendoza", "Rosario", "La Plata",
	"Mar del Plata", "Salta", "Tucuman", "Bariloche", "Ushuaia",
}

var streets = []string{
	"Av. Libertador", "Calle Florida", "Av. Corrientes", "Av. 9 de Julio",
	"Av. Santa Fe", "Calle Defensa", "Calle Lavalle", "Av. de Mayo",
}

var categories = []categoryDef{
	{
		Name: "Electronics",
		StoreNames: []string{
			"Lauti Tech", "ByteHaven", "PixelForge", "CircuitWorks",
			"VoltStreet", "GadgetHub", "NeonCircuit", "ChipBazaar",
		},
		StoreDescriptions: []string{
			"Quality electronics, accessories and smart gadgets.",
			"Your one-stop shop for everything tech.",
			"Curated electronics from trusted brands.",
		},
		Products: []productTemplate{
			{"Wireless Headphones", "wireless headphones", "Over-ear noise-cancelling headphones with long battery life.", 40, 220},
			{"Bluetooth Speaker", "bluetooth speaker", "Portable speaker with rich sound and waterproof design.", 25, 180},
			{"Mechanical Keyboard", "mechanical keyboard", "Tactile mechanical keyboard with hot-swappable switches.", 60, 200},
			{"Gaming Mouse", "gaming mouse", "High-precision wired mouse with customizable buttons.", 20, 120},
			{"USB-C Hub", "usb hub", "Multi-port hub with HDMI, USB-A and SD card support.", 25, 90},
			{"Webcam 1080p", "webcam", "Full HD webcam with autofocus and built-in microphone.", 30, 130},
			{"Smartwatch", "smartwatch", "Fitness-tracking smartwatch with heart rate monitor.", 80, 350},
			{"Wireless Charger", "wireless charger", "Fast wireless charging pad compatible with most phones.", 15, 60},
			{"LED Desk Lamp", "desk lamp", "Adjustable LED desk lamp with USB charging port.", 20, 80},
			{"Portable SSD", "ssd", "High-speed external SSD with USB-C connectivity.", 50, 250},
		},
	},
	{
		Name: "Garden",
		StoreNames: []string{
			"Green Garden Supply", "Verde Botanical", "EarthRoot",
			"BloomYard", "Wildleaf Co", "FernHaven", "Patio and Petal",
		},
		StoreDescriptions: []string{
			"Everything you need to grow a beautiful garden.",
			"Tools, seeds and supplies for plant lovers.",
			"Sustainable gardening essentials sourced with care.",
		},
		Products: []productTemplate{
			{"Organic Fertilizer", "fertilizer", "All-natural fertilizer for vegetables, flowers and shrubs.", 10, 40},
			{"Pruning Shears", "pruning shears", "Stainless steel bypass pruners with ergonomic grip.", 12, 50},
			{"Self-Watering Planter", "planter pot", "Modern planter with built-in water reservoir.", 15, 70},
			{"Garden Hose 25ft", "garden hose", "Flexible reinforced hose with brass fittings.", 20, 80},
			{"Heirloom Seeds Pack", "seeds packet", "Non-GMO open-pollinated heirloom variety seeds.", 3, 15},
			{"Garden Gloves", "gardening gloves", "Durable gloves with reinforced fingertips.", 6, 25},
			{"Hand Trowel", "garden trowel", "Lightweight hand trowel for digging and transplanting.", 8, 30},
			{"Watering Can 5L", "watering can", "Galvanized steel watering can with long spout.", 14, 45},
			{"Compost Bin", "compost bin", "Odor-controlled kitchen compost bin.", 25, 90},
			{"Bamboo Plant Stakes", "plant stakes", "Pack of natural bamboo stakes for climbing plants.", 5, 20},
		},
	},
	{
		Name: "Books",
		StoreNames: []string{
			"Urban Bookshelf", "Inkwell and Co", "PageTurner",
			"Margin Press", "Foliant Books", "ChapterOne",
		},
		StoreDescriptions: []string{
			"A curated selection of books and stationery.",
			"Reading and writing essentials for every desk.",
			"Independent bookshop with handpicked titles.",
		},
		Products: []productTemplate{
			{"Leather Journal", "leather journal", "Handcrafted leather journal with acid-free paper.", 15, 60},
			{"Bamboo Book Stand", "book stand", "Adjustable book stand for hands-free reading.", 12, 45},
			{"Highlighter Set", "highlighters", "Chisel-tip highlighters in vibrant colors.", 5, 20},
			{"Fountain Pen", "fountain pen", "Smooth-writing fountain pen with refillable converter.", 25, 150},
			{"Notebook A5", "notebook", "Dot-grid notebook with soft-touch cover.", 8, 30},
			{"Reading Light", "reading light", "Clip-on LED reading light with adjustable arm.", 10, 40},
			{"Bookmark Set", "bookmark", "Set of metal bookmarks with elegant designs.", 4, 18},
			{"Desk Organizer", "desk organizer", "Wooden desk organizer with multiple compartments.", 18, 70},
			{"Watercolor Set", "watercolor paint", "Travel watercolor set with brush and palette.", 15, 80},
		},
	},
	{
		Name: "Kitchen",
		StoreNames: []string{
			"Copper and Clay", "HearthGoods", "Saucepan Society",
			"The Whisk Room", "Larder Lane", "Brass Spoon Co",
		},
		StoreDescriptions: []string{
			"Cookware and kitchen tools for the home chef.",
			"Quality kitchen essentials, sourced with care.",
			"Everything you need for cooking and entertaining.",
		},
		Products: []productTemplate{
			{"Chef's Knife", "chef knife", "High-carbon steel chef's knife with riveted handle.", 25, 200},
			{"Cutting Board", "cutting board", "Hardwood cutting board with juice groove.", 20, 90},
			{"Cast Iron Skillet", "cast iron skillet", "Pre-seasoned cast iron skillet, oven-safe.", 30, 120},
			{"Mixing Bowl Set", "mixing bowls", "Stainless steel nesting mixing bowls.", 25, 80},
			{"Wooden Spatula", "wooden spatula", "Hand-carved wooden spatula, gentle on cookware.", 6, 25},
			{"Linen Apron", "kitchen apron", "Adjustable linen apron with front pocket.", 18, 60},
			{"French Press", "french press", "Borosilicate glass French press coffee maker.", 22, 80},
			{"Spice Rack", "spice rack", "Bamboo spice rack with labeled jars.", 25, 90},
			{"Pasta Maker", "pasta maker", "Manual pasta machine with adjustable thickness.", 40, 180},
		},
	},
	{
		Name: "Fashion",
		StoreNames: []string{
			"Threadwood", "Atelier Norte", "Linen and Loom",
			"Kindred Goods", "WoolWeft", "MarchStreet",
		},
		StoreDescriptions: []string{
			"Timeless apparel and accessories.",
			"Independent fashion for everyday wear.",
			"Curated wardrobe staples and accessories.",
		},
		Products: []productTemplate{
			{"Cotton T-Shirt", "tshirt", "Soft organic cotton tee in classic fit.", 15, 60},
			{"Wool Beanie", "beanie hat", "Cozy merino wool beanie in solid colors.", 18, 55},
			{"Leather Wallet", "leather wallet", "Slim leather wallet with RFID protection.", 25, 120},
			{"Canvas Backpack", "canvas backpack", "Durable canvas backpack with padded laptop sleeve.", 35, 150},
			{"Sunglasses", "sunglasses", "Polarized sunglasses with UV protection.", 25, 180},
			{"Knit Scarf", "knit scarf", "Hand-knit scarf in soft alpaca blend.", 20, 90},
			{"Leather Belt", "leather belt", "Full-grain leather belt with brass buckle.", 30, 120},
			{"Field Watch", "wristwatch", "Minimalist field watch with leather strap.", 60, 300},
		},
	},
	{
		Name: "Home",
		StoreNames: []string{
			"Hearth and Hue", "Maison Quinta", "Loft Goods",
			"Quietude Home", "Soft Habit", "DwellRoom",
		},
		StoreDescriptions: []string{
			"Home goods that bring warmth to any space.",
			"Furniture, lighting and decor for modern living.",
			"Curated objects for the well-lived home.",
		},
		Products: []productTemplate{
			{"Linen Cushion", "cushion", "Stonewashed linen cushion cover with hidden zip.", 20, 80},
			{"Throw Blanket", "throw blanket", "Soft woven throw in neutral tones.", 30, 140},
			{"Ceramic Vase", "ceramic vase", "Handmade stoneware vase with matte finish.", 25, 120},
			{"Soy Candle", "scented candle", "Hand-poured soy candle with essential oils.", 15, 55},
			{"Wall Mirror", "wall mirror", "Round mirror with brass frame.", 40, 200},
			{"Rattan Basket", "rattan basket", "Woven rattan storage basket with handles.", 25, 90},
			{"Floor Lamp", "floor lamp", "Tripod floor lamp with linen shade.", 60, 280},
			{"Photo Frame Set", "photo frame", "Set of three matching wooden photo frames.", 18, 75},
		},
	},
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file: %v", err)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

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
	if err := resp.Body.Close(); err != nil {
		log.Printf("closing health check body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Health check returned unexpected status %d", resp.StatusCode)
	}
	log.Println("Server is reachable. Starting seed...")

	suffix := fmt.Sprintf("%05d", rng.Intn(100000))
	log.Printf("Run suffix: %s", suffix)

	sellers := generateSellers(rng, suffix)
	log.Printf("Generated %d random sellers for this run", len(sellers))

	var totalStores, totalProducts, failedStores, failedProducts int

	for _, seller := range sellers {
		accountID, ok := registerStore(client, *baseURL, &seller)
		if !ok {
			failedStores++
			continue
		}
		totalStores++

		for _, product := range seller.Products {
			if createProduct(client, *baseURL, accountID, &product, unsplashKey) {
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

// generateSellers builds a randomized list of 3-5 sellers, each from a distinct category.
func generateSellers(rng *rand.Rand, suffix string) []sellerSeed {
	n := 3 + rng.Intn(3) // 3, 4 or 5
	if n > len(categories) {
		n = len(categories)
	}

	perm := rng.Perm(len(categories))[:n]
	out := make([]sellerSeed, 0, n)
	for i, idx := range perm {
		out = append(out, generateSeller(rng, &categories[idx], suffix, i))
	}
	return out
}

// generateSeller picks a random store name, address and 6-9 products from the category pool.
func generateSeller(rng *rand.Rand, cat *categoryDef, suffix string, index int) sellerSeed {
	storeName := cat.StoreNames[rng.Intn(len(cat.StoreNames))]
	description := cat.StoreDescriptions[rng.Intn(len(cat.StoreDescriptions))]

	slug := strings.ToLower(storeName)
	slug = strings.ReplaceAll(slug, " ", ".")
	email := fmt.Sprintf("%s+%s%d@example.com", slug, suffix, index)

	address := fmt.Sprintf("%s %d, %s",
		streets[rng.Intn(len(streets))],
		100+rng.Intn(9000),
		cities[rng.Intn(len(cities))],
	)
	phone := fmt.Sprintf("555-%04d", rng.Intn(10000))

	productCount := 6 + rng.Intn(4) // 6, 7, 8 or 9
	products := generateProducts(rng, cat, productCount)

	return sellerSeed{
		Email:       email,
		Password:    "securepass1",
		Name:        storeName,
		Description: description,
		Address:     address,
		PhoneNumber: phone,
		Products:    products,
	}
}

// generateProducts picks N distinct templates from the category and decorates each
// with a random adjective + model code so names rarely collide across runs.
func generateProducts(rng *rand.Rand, cat *categoryDef, n int) []productSeed {
	if n > len(cat.Products) {
		n = len(cat.Products)
	}
	perm := rng.Perm(len(cat.Products))[:n]
	out := make([]productSeed, 0, n)
	for _, idx := range perm {
		tmpl := cat.Products[idx]
		adj := productAdjectives[rng.Intn(len(productAdjectives))]
		model := modelCodes[rng.Intn(len(modelCodes))]
		name := fmt.Sprintf("%s %s %s", adj, tmpl.BaseName, model)

		price := tmpl.MinPrice + rng.Float64()*(tmpl.MaxPrice-tmpl.MinPrice)
		price = float64(int(price)) + 0.99 // round to .99

		out = append(out, productSeed{
			Name:        name,
			Description: tmpl.Description,
			Category:    cat.Name,
			Stock:       10 + rng.Intn(91), // 10..100
			Price:       price,
			ImageQuery:  tmpl.ImageQuery,
		})
	}
	return out
}

// registerStore POSTs to /auth/register/store and returns the account_id on success.
func registerStore(client *http.Client, baseURL string, seller *sellerSeed) (string, bool) {
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
		accountID, ok := body["account_id"].(string)
		if !ok {
			log.Printf("[store] WARNING: account_id missing in response for %q", seller.Name)
		}
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
// patches the image_url via PUT if Unsplash returns a random photo for the query.
func createProduct(client *http.Client, baseURL, storeID string, product *productSeed, unsplashKey string) bool {
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

	productID, ok := body["id"].(string)
	if !ok {
		log.Printf("[product] WARNING: id missing in response for %q", product.Name)
	}
	log.Printf("[product] Created %q (id=%s)", product.Name, productID)

	if unsplashKey == "" || productID == "" {
		return true
	}

	imageURL, err := fetchUnsplashURL(client, unsplashKey, product.ImageQuery)
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

// fetchUnsplashURL queries the Unsplash Random Photo API and returns the URL of
// a random photo matching the query. Each call returns a different photo even
// for the same query, which is what gives the seed its visual variety.
func fetchUnsplashURL(client *http.Client, accessKey, query string) (string, error) {
	apiURL := "https://api.unsplash.com/photos/random?query=" + url.QueryEscape(query) + "&orientation=squarish"

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating unsplash request: %w", err)
	}
	req.Header.Set("Authorization", "Client-ID "+accessKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling unsplash API: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("closing response body: %v", closeErr)
		}
	}()

	// Unsplash returns 404 when no photo matches the query.
	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unsplash returned status %d", resp.StatusCode)
	}

	var apiResp struct {
		URLs struct {
			Regular string `json:"regular"`
		} `json:"urls"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("decoding unsplash response: %w", err)
	}
	return apiResp.URLs.Regular, nil
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
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("closing response body: %v", closeErr)
		}
	}()

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
