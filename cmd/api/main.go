package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/LautaroBlasco23/lauti-market-backend/database"
	apiInfrastructure "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	authinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure"
	authUtils "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/utils"
	imageinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/image/infrastructure"
	productinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/product/infrastructure"
	storeinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure"
	userinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	postgres, err := database.NewPostgres(database.PostgresConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "lauti_market"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	})
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer func() {
		if err := postgres.Close(); err != nil {
			log.Printf("error closing database connection: %v", err)
		}
	}()

	db := postgres.DB()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	uuidGen := apiInfrastructure.NewUUIDGenerator()

	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-in-production")
	jwtGen := authUtils.NewJWTGenerator(jwtSecret, 24*time.Hour)
	validate := func(token string) (string, error) {
		claims, err := jwtGen.Validate(token)
		if err != nil {
			return "", err
		}
		return claims.AccountID, nil
	}

	userModule := userinfra.Wire(mux, db, uuidGen, validate)
	storeModule := storeinfra.Wire(mux, db, uuidGen)

	authinfra.Wire(mux, db, uuidGen, userModule, storeModule, authUtils.JwtConfig{
		JWTSecret:     jwtSecret,
		JWTExpiration: 24 * time.Hour,
	})

	imageModule, err := imageinfra.Wire(getEnv("IMAGE_STORE_ADDR", "localhost:50051"))
	if err != nil {
		log.Fatalf("connecting to image service: %v", err)
	}
	defer imageModule.Close()

	productinfra.Wire(mux, db, uuidGen, storeModule.Repository, imageModule.Client)

	server := &http.Server{
		Addr:         getEnv("PORT", ":8000"),
		Handler:      apiInfrastructure.CORSMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("Shutting down server...")
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}
	return fallback
}
