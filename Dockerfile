# --- build stage ---
FROM golang:1.25-alpine AS builder
WORKDIR /app

# deps
COPY go.mod go.sum ./
RUN go mod download

# build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -o app ./cmd/api && \
  go build -ldflags="-s -w" -o healthcheck ./cmd/healthcheck

# --- runtime stage ---
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/app /app/app
COPY --from=builder /app/healthcheck /app/healthcheck
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/app"]
