#!/usr/bin/env bash
set -euo pipefail

# Defaults
PRODUCTS=20
STORES=2
BASE_URL="http://localhost:8000"

# Parse CLI args
for arg in "$@"; do
  case "$arg" in
    --products=*) PRODUCTS="${arg#*=}" ;;
    --stores=*)   STORES="${arg#*=}" ;;
    --base-url=*) BASE_URL="${arg#*=}" ;;
    --help)
      echo "Usage: $0 [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  --products=N       Number of products to create (default: 20)"
      echo "  --stores=N         Number of store accounts to register (default: 2)"
      echo "  --base-url=URL     API base URL (default: http://localhost:8000)"
      echo "  --help             Print this help and exit"
      exit 0
      ;;
    *) echo "Unknown argument: $arg"; exit 1 ;;
  esac
done

# Check dependencies
for dep in curl jq; do
  if ! command -v "$dep" &>/dev/null; then
    echo "Error: '$dep' is required but not installed. Please install it and retry."
    exit 1
  fi
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGES_DIR="$SCRIPT_DIR/fake-data-images"

# Collect images
IMAGES=()
if [[ -d "$IMAGES_DIR" ]]; then
  while IFS= read -r -d '' f; do
    IMAGES+=("$f")
  done < <(find "$IMAGES_DIR" -maxdepth 1 -type f \( -iname "*.jpg" -o -iname "*.jpeg" -o -iname "*.png" -o -iname "*.webp" \) -print0 2>/dev/null)
fi

IMAGE_COUNT=${#IMAGES[@]}
if [[ $IMAGE_COUNT -gt 0 ]]; then
  echo "📷 Found $IMAGE_COUNT image(s) in $IMAGES_DIR"
else
  echo "📷 No images found — products will be created without images"
fi

# ─── Fake Data ───────────────────────────────────────────────────────────────

STORE_NAMES=("TechWorld" "StyleHub" "FreshMarket" "HomeNest" "SportZone")
STORE_DESCS=(
  "Your one-stop shop for the latest technology and gadgets"
  "Trendy clothing and accessories for every occasion"
  "Fresh organic food and artisan products delivered to your door"
  "Beautiful home decor and essentials for modern living"
  "Everything you need for an active and healthy lifestyle"
)
STORE_ADDRS=(
  "123 Main St, Buenos Aires"
  "456 Fashion Ave, Córdoba"
  "789 Green Rd, Rosario"
  "321 Oak Blvd, Mendoza"
  "654 Sport Way, La Plata"
)
STORE_PHONES=(
  "+5491112345678"
  "+5493516789012"
  "+5493413456789"
  "+5492614567890"
  "+5492215678901"
)

CATEGORIES=("electronics" "clothing" "food" "home" "sports" "books" "toys" "beauty")

declare -A PRODUCT_NAMES
PRODUCT_NAMES[electronics]="Wireless Headphones|Bluetooth Speaker|USB-C Hub|Mechanical Keyboard|Smart Watch|Portable Charger|LED Monitor|Webcam"
PRODUCT_NAMES[clothing]="Cotton T-Shirt|Denim Jacket|Running Shoes|Wool Sweater|Casual Shorts|Rain Jacket|Leather Belt|Knit Hat"
PRODUCT_NAMES[food]="Organic Coffee Beans|Dark Chocolate Bar|Extra Virgin Olive Oil|Artisan Honey|Mixed Nuts Bag|Herbal Tea Set|Granola Mix|Hot Sauce"
PRODUCT_NAMES[home]="Ceramic Vase|LED Desk Lamp|Bamboo Cutting Board|Scented Candle|Storage Basket|Wall Clock|Plant Pot|Throw Pillow"
PRODUCT_NAMES[sports]="Yoga Mat|Resistance Bands|Jump Rope|Foam Roller|Water Bottle|Gym Gloves|Ankle Weights|Stretching Strap"
PRODUCT_NAMES[books]="Programming Guide|Science Fiction Novel|Cookbook|Self-Help Journal|History Atlas|Children Storybook|Art Sketchbook|Poetry Collection"
PRODUCT_NAMES[toys]="Building Blocks Set|Remote Control Car|Puzzle Box|Stuffed Animal|Board Game|Art Supply Kit|Model Rocket|Sand Play Set"
PRODUCT_NAMES[beauty]="Moisturizing Cream|Organic Shampoo|Sunscreen SPF 50|Lip Balm Set|Face Serum|Nail Polish Kit|Rose Water Toner|Exfoliating Scrub"

declare -A PRODUCT_DESCS
PRODUCT_DESCS[electronics]="High-quality audio device with noise cancellation and long battery life|Compact speaker with 360-degree surround sound and water resistance|Multi-port hub supporting fast charging and high-speed data transfer|Tactile keys with RGB backlight and programmable macros|Tracks fitness, notifications, and heart rate around the clock|10000mAh bank with fast charging for all your devices|Crisp display with adjustable brightness and blue-light filter|Crystal-clear 1080p video for remote work and streaming"
PRODUCT_DESCS[clothing]="Soft breathable fabric perfect for everyday casual wear|Classic cut with durable stitching and timeless style|Lightweight and responsive sole for all-terrain comfort|Warm merino blend that keeps you cozy in cold weather|Relaxed fit with deep pockets for warm-weather adventures|Waterproof shell that keeps you dry in any storm|Full-grain leather with classic buckle and clean finish|Ribbed knit construction to keep your head warm all winter"
PRODUCT_DESCS[food]="Single-origin beans roasted to a smooth medium profile|Rich cacao blend with 72 percent dark chocolate content|Cold-pressed from hand-picked olives for superior flavor|Raw wildflower honey harvested from sustainable apiaries|Premium selection of cashews, almonds, and walnuts|Calming blend of chamomile, mint, and lavender leaves|Rolled oats with seeds and dried fruit for a hearty breakfast|Fermented small-batch sauce with a bold smoky kick"
PRODUCT_DESCS[home]="Hand-thrown stoneware with a minimalist matte finish|Energy-efficient lamp with adjustable arm and warm glow|Eco-friendly board with natural antibacterial properties|Long-lasting soy wax candle with calming essential oils|Woven seagrass basket ideal for organizing any room|Silent quartz movement with a clean Nordic dial design|Terracotta pot with drainage hole for healthy root growth|Hypoallergenic fill with a soft velvet cover for the sofa"
PRODUCT_DESCS[sports]="Non-slip surface with extra cushioning for joint support|Set of five resistance levels for strength and rehab work|Lightweight speed rope with comfortable foam handles|Dense foam cylinder for post-workout muscle recovery|Double-walled insulated bottle keeps drinks cold for 24 hours|Padded palm protection for grip during heavy lifts|Adjustable strap weights for low-impact toning exercises|Extra-long band for full-body stretching and mobility work"
PRODUCT_DESCS[books]="Comprehensive guide covering algorithms and system design patterns|Epic space adventure that redefines the boundaries of imagination|Step-by-step recipes from seasonal ingredients for every skill level|Guided prompts and blank pages to track goals and reflections|Illustrated timeline of world events from ancient to modern times|Whimsical tale with vibrant illustrations for kids aged three to eight|High-quality paper suitable for pencil, ink, and watercolor|Curated verses that explore love, loss, and the human condition"
PRODUCT_DESCS[toys]="Colorful interlocking bricks that spark creativity for all ages|Full-function remote car with rechargeable battery and turbo mode|Laser-cut wooden puzzle with 500 pieces and a satisfying finish|Ultra-soft plush friend made from hypoallergenic materials|Strategy game for two to four players with easy-to-learn rules|Complete set of washable markers, colored pencils, and crayons|Build-and-launch kit with real igniter for outdoor use|Fine-grain sand with molds and tools for sensory play"
PRODUCT_DESCS[beauty]="Deeply hydrating formula with hyaluronic acid and vitamin E|Sulfate-free shampoo with argan oil for shine and softness|Broad-spectrum protection with lightweight non-greasy texture|Nourishing set of five flavored balms for soft and smooth lips|Vitamin C brightening serum that visibly reduces dark spots|Chip-resistant formula in 12 seasonal shades with quick dry|Alcohol-free toner that soothes and balances the skin barrier|Gentle exfoliator with walnut shell powder for a radiant glow"

declare -A PRICE_RANGES
PRICE_RANGES[electronics]="2999 99999"
PRICE_RANGES[clothing]="1499 19999"
PRICE_RANGES[food]="299 4999"
PRICE_RANGES[home]="999 29999"
PRICE_RANGES[sports]="999 14999"
PRICE_RANGES[books]="799 4999"
PRICE_RANGES[toys]="999 8999"
PRICE_RANGES[beauty]="499 7999"

SUFFIXES=("Black" "White" "Silver" "Blue" "Red" "Green" "XS" "S" "M" "L" "XL" "Pro" "Plus" "Lite" "v2" "Mini")

random_int() {
  local min=$1 max=$2
  echo $(( min + RANDOM % (max - min + 1) ))
}

random_element() {
  local arr=("$@")
  echo "${arr[$(random_int 0 $(( ${#arr[@]} - 1 )))]}"
}

# ─── Register Stores ─────────────────────────────────────────────────────────

echo ""
echo "🏪 Registering $STORES store(s)..."

STORE_IDS=()
MAX_STORE_IDX=$(( ${#STORE_NAMES[@]} - 1 ))

for (( i=0; i<STORES; i++ )); do
  IDX=$(( i % (MAX_STORE_IDX + 1) ))
  NAME="${STORE_NAMES[$IDX]}"
  DESC="${STORE_DESCS[$IDX]}"
  ADDR="${STORE_ADDRS[$IDX]}"
  PHONE="${STORE_PHONES[$IDX]}"
  EMAIL="store-$(echo "$NAME" | tr '[:upper:]' '[:lower:]')-$RANDOM@fakedata.local"

  PAYLOAD=$(jq -n \
    --arg email "$EMAIL" \
    --arg name "$NAME" \
    --arg desc "$DESC" \
    --arg addr "$ADDR" \
    --arg phone "$PHONE" \
    '{email: $email, password: "fakepassword123", name: $name, description: $desc, address: $addr, phone_number: $phone}')

  RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register/store" \
    -H "Content-Type: application/json" \
    -d "$PAYLOAD" 2>/dev/null) || {
    echo "❌ API not reachable at $BASE_URL. Start with 'make dev' first."
    exit 1
  }

  HTTP_CODE=$(echo "$RESPONSE" | tail -1)
  BODY=$(echo "$RESPONSE" | head -n -1)

  if [[ "$HTTP_CODE" != "201" && "$HTTP_CODE" != "200" ]]; then
    echo "❌ Failed to register store '$NAME' (HTTP $HTTP_CODE): $BODY"
    exit 1
  fi

  STORE_ID=$(echo "$BODY" | jq -r '.account_id // .id // empty')

  if [[ -z "$STORE_ID" ]]; then
    echo "❌ Could not parse store ID from response: $BODY"
    exit 1
  fi

  # Register doesn't return a token — login to get one
  LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"fakepassword123\"}" 2>/dev/null)
  LOGIN_CODE=$(echo "$LOGIN_RESPONSE" | tail -1)
  LOGIN_BODY=$(echo "$LOGIN_RESPONSE" | head -n -1)

  if [[ "$LOGIN_CODE" != "200" ]]; then
    echo "❌ Failed to login as '$NAME' (HTTP $LOGIN_CODE): $LOGIN_BODY"
    exit 1
  fi

  TOKEN=$(echo "$LOGIN_BODY" | jq -r '.token // empty')

  if [[ -z "$TOKEN" ]]; then
    echo "❌ Could not parse token from login response: $LOGIN_BODY"
    exit 1
  fi

  STORE_IDS+=("$STORE_ID:$TOKEN")
  echo "  ✅ Registered '$NAME' (id: $STORE_ID)"
done

# ─── Create Products ─────────────────────────────────────────────────────────

echo ""
echo "📦 Creating $PRODUCTS product(s) across ${#STORE_IDS[@]} store(s)..."

SUCCESS=0
FAIL=0

for (( i=0; i<PRODUCTS; i++ )); do
  # Pick random store
  STORE_ENTRY=$(random_element "${STORE_IDS[@]}")
  STORE_ID="${STORE_ENTRY%%:*}"
  TOKEN="${STORE_ENTRY#*:}"

  # Pick random category
  CATEGORY=$(random_element "${CATEGORIES[@]}")

  # Pick random name + description
  IFS='|' read -ra NAMES <<< "${PRODUCT_NAMES[$CATEGORY]}"
  IFS='|' read -ra DESCS <<< "${PRODUCT_DESCS[$CATEGORY]}"
  NAME_IDX=$(random_int 0 $(( ${#NAMES[@]} - 1 )))
  PRODUCT_NAME="${NAMES[$NAME_IDX]} - $(random_element "${SUFFIXES[@]}")"
  PRODUCT_DESC="${DESCS[$NAME_IDX]}"

  # Price (convert cents to decimal)
  read -ra RANGE <<< "${PRICE_RANGES[$CATEGORY]}"
  PRICE_CENTS=$(random_int "${RANGE[0]}" "${RANGE[1]}")
  PRICE=$(awk "BEGIN {printf \"%.2f\", $PRICE_CENTS / 100}")

  STOCK=$(random_int 1 200)

  # Build curl command
  CURL_ARGS=(
    -s -w "\n%{http_code}"
    -X POST "$BASE_URL/stores/$STORE_ID/products"
    -H "Authorization: Bearer $TOKEN"
    -F "name=$PRODUCT_NAME"
    -F "description=$PRODUCT_DESC"
    -F "category=$CATEGORY"
    -F "stock=$STOCK"
    -F "price=$PRICE"
  )

  # ~70% chance to attach image if available
  if [[ $IMAGE_COUNT -gt 0 && $(random_int 1 10) -le 7 ]]; then
    IMG=$(random_element "${IMAGES[@]}")
    CURL_ARGS+=(-F "image=@$IMG")
  fi

  RESPONSE=$(curl "${CURL_ARGS[@]}" 2>/dev/null)
  HTTP_CODE=$(echo "$RESPONSE" | tail -1)
  BODY=$(echo "$RESPONSE" | head -n -1)

  if [[ "$HTTP_CODE" == "201" || "$HTTP_CODE" == "200" ]]; then
    SUCCESS=$(( SUCCESS + 1 ))
    printf "  [%d/%d] ✅ %-45s (%s, \$%s, stock: %d)\n" \
      "$(( i + 1 ))" "$PRODUCTS" "$PRODUCT_NAME" "$CATEGORY" "$PRICE" "$STOCK"
  else
    FAIL=$(( FAIL + 1 ))
    printf "  [%d/%d] ⚠️  Failed '%s' (HTTP %s): %s\n" \
      "$(( i + 1 ))" "$PRODUCTS" "$PRODUCT_NAME" "$HTTP_CODE" "$BODY"
  fi
done

# ─── Summary ─────────────────────────────────────────────────────────────────

echo ""
echo "─────────────────────────────────────────"
echo "✅ Created $SUCCESS/$PRODUCTS products across ${#STORE_IDS[@]} store(s)"
[[ $FAIL -gt 0 ]] && echo "⚠️  $FAIL product(s) failed"
echo "─────────────────────────────────────────"
