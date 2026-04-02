#!/usr/bin/env bash
set -euo pipefail

ENV_FILE=".env"
EXAMPLE_FILE=".env.example"

SECURE_FIELDS=("DB_PASSWORD" "JWT_SECRET")
MANUAL_FIELDS=("UNSPLASH_ACCESS_KEY" "MERCADO_PAGO_ACCESS_TOKEN" "VITE_MERCADO_PAGO_PUBLIC_KEY")

generate_secret() {
    openssl rand -hex 32
}

is_in_array() {
    local needle="$1"; shift
    local item
    for item; do [[ "$item" == "$needle" ]] && return 0; done
    return 1
}

if [ ! -f "$EXAMPLE_FILE" ]; then
    echo "Error: $EXAMPLE_FILE not found. Run this script from the project root." >&2
    exit 1
fi

echo ""
echo "Setting up .env..."
echo ""

[ -f "$ENV_FILE" ] || touch "$ENV_FILE"

added=0; skipped=0; manual=0

while IFS= read -r line || [ -n "$line" ]; do
    # Skip blank lines and comments
    [[ -z "${line//[[:space:]]/}" ]] && continue
    [[ "$line" =~ ^[[:space:]]*# ]] && continue

    key="${line%%=*}"

    if grep -q "^${key}=" "$ENV_FILE" 2>/dev/null; then
        printf "  skip   %s\n" "$key"
        (( skipped++ )) || true
        continue
    fi

    if is_in_array "$key" "${SECURE_FIELDS[@]}"; then
        secret=$(generate_secret)
        printf '%s="%s"\n' "$key" "$secret" >> "$ENV_FILE"
        printf "  gen    %s\n" "$key"
        (( added++ )) || true
    elif is_in_array "$key" "${MANUAL_FIELDS[@]}"; then
        echo "$line" >> "$ENV_FILE"
        printf "  warn   %s  (placeholder — update manually)\n" "$key"
        (( manual++ )) || true
    else
        echo "$line" >> "$ENV_FILE"
        printf "  set    %s\n" "$key"
        (( added++ )) || true
    fi
done < "$EXAMPLE_FILE"

echo ""
echo "Done: $added added, $skipped skipped, $manual need manual update."

if [ "$manual" -gt 0 ]; then
    echo ""
    echo "Update these keys in .env before running the app:"
    for field in "${MANUAL_FIELDS[@]}"; do
        if grep -q "^${field}=" "$ENV_FILE" 2>/dev/null; then
            printf "  - %s\n" "$field"
        fi
    done
fi

echo ""
